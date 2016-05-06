package processor

import (
  "fmt"
  "image"

  "github.com/benparham/lyon/challenge/common/imagedata"
  "github.com/benparham/lyon/challenge/common/syncdata"
  chromath "github.com/jkl1337/go-chromath"
  converter "github.com/benparham/lyon/challenge/colorconverter"
)

const PROCESSOR_BUFFER = 200

const BOUNDS_FRACTION = 0.5   // Fraction of img bounds to check
const BOUNDS_CUTOFF = 50      // Don't shrink bounds if img size is less than this

const MAX_COLORS = 3
const PX_THRESHOLD_RATIO = 0.25
const PX_THRESHOLD_RATIO_WHITE = 0.5

type processorError struct { err error }
func (pe processorError) Error() string {
  return fmt.Sprintf("[Processor] %s", pe.err.Error())
}

type Processor struct {
  inChan imagedata.DataChannel
  *syncdata.SyncData
  outChan imagedata.DataChannel
  multiId string
  whiteId string
  buckets []colorBucket
}

func New(
  colorPath string,
  inChan imagedata.DataChannel,
  syncData *syncdata.SyncData,
  ) (*Processor, error) {

  bkts, mId, wId, err := loadColorBuckets(colorPath)
  if err != nil { return nil, err }

  return &Processor{
    inChan: inChan,
    SyncData: syncData,
    outChan: make(imagedata.DataChannel, PROCESSOR_BUFFER),
    multiId: mId,
    whiteId: wId,
    buckets: bkts,
  }, nil
}

func (p *Processor) send(data *imagedata.ImageData) {
  p.outChan <- data
}

func (p *Processor) sendErr(err error) {
  if err == nil { return }
  p.ErrChan <- processorError{ err }
}

func (p *Processor) finish() {
  close(p.outChan)
  p.Wg.Done()
}

func (p *Processor) Run() imagedata.DataChannel {
  go func() {
    defer p.finish()
    p.process()
  }()

  return p.outChan
}

func (p *Processor) runWorker(imgData *imagedata.ImageData) {
  img := *imgData.Data
  bounds := calcInnerBounds(img.Bounds())

  // Tally a count for each bucket
  counter := newBucketCounter(len(p.buckets), MAX_COLORS)

  // Iterate over the inner bounds and grab pixels
  for x := bounds.Min.X; x < bounds.Max.X; x++ {
    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
      labColor := converter.RgbaToLab(img.At(x, y))
      matchIdx, _ := p.matchLabColor(labColor)
      counter.increment(matchIdx)
    }
  }


  topColors := counter.getTop()

  // Minimum px count required to be included
  numPx := (bounds.Max.X - bounds.Min.X) * (bounds.Max.Y - bounds.Min.Y)
  pxThreshold := int(float64(numPx) * PX_THRESHOLD_RATIO)
  pxThresholdWhite := int(float64(numPx) * PX_THRESHOLD_RATIO_WHITE)

  for _, result := range topColors {
    if result.count < pxThreshold { break }
    bucket := p.buckets[result.idx]

    // Stricter threshold for white px
    if bucket.colorId == p.whiteId && result.count < pxThresholdWhite { break }

    imgData.ColorIds = append(imgData.ColorIds, bucket.colorId)
  }

  // Multi if no bucket counts broke the threshold
  if len(imgData.ColorIds) == 0 {
    imgData.ColorIds = append(imgData.ColorIds, p.multiId)
  }

  imgData.Data = nil
  p.send(imgData)
}

// Get a central inner portion of the total image
func calcInnerBounds(bounds image.Rectangle) image.Rectangle {
  width := bounds.Max.X - bounds.Min.X
  height := bounds.Max.Y - bounds.Min.Y

  if width < BOUNDS_CUTOFF || height < BOUNDS_CUTOFF {
    return bounds
  }

  newWidth := int(float64(width) * BOUNDS_FRACTION)
  newHeight := int(float64(height) * BOUNDS_FRACTION)

  xOffset := int((width - newWidth) / 2)
  yOffset := int((height - newHeight) / 2)

  newMin := image.Point{X: bounds.Min.X + xOffset, Y: bounds.Min.Y + yOffset}
  newMax := image.Point{X: newMin.X + newWidth, Y: newMin.Y + newHeight}

  return image.Rectangle{Min: newMin, Max: newMax}
}

func (p *Processor) process() {
  for imgData := range p.inChan {
    p.runWorker(imgData)
  }
}

// Return index into array of color buckets and dist value of matching color
func (p *Processor) matchLabColor(labColor chromath.Lab) (int, float64) {
  minIdx := 0
  minDist := converter.CalcDist(labColor, p.buckets[0].labColor)

  for i := 1; i < len(p.buckets); i++ {
    newDist := converter.CalcDist(labColor, p.buckets[i].labColor)
    if newDist < minDist {
      minIdx = i
      minDist = newDist
    }
  }

  return minIdx, minDist
}
