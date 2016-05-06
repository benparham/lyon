package exporter

import (
  "fmt"
  "os"

  "github.com/benparham/lyon/challenge/common/imagedata"
  "github.com/benparham/lyon/challenge/common/syncdata"
)

const EXPORTER_BUFFER = 200

// ========= Custom Error
type exportError struct { err error }
func (ee exportError) Error() string {
  return fmt.Sprintf("[Exporter] %s", ee.err.Error())
}

// ========= Out Buffer
// will buffer data in memory before writing to file
type outBuffer struct {
  outFile *os.File
  buffer []*imagedata.ImageData
}

func newOutBuffer(outPath string) (*outBuffer, error) {
  f, err := os.Create(outPath)
  if err != nil { return nil, err }

  return &outBuffer{
    outFile: f,
    buffer: make([]*imagedata.ImageData, 0, EXPORTER_BUFFER),
  }, nil
}

func (ob *outBuffer) write(data *imagedata.ImageData) error {
  var err error
  if len(ob.buffer) == cap(ob.buffer) {
    err = ob.flush()
  }
  ob.buffer = append(ob.buffer, data)
  return err
}

func (ob *outBuffer) flush() error {
  var err error
L:
  for _, data := range ob.buffer {
    for _, colorId := range data.ColorIds {
      _, err = ob.outFile.WriteString(fmt.Sprintf("%s,%s\n", data.Id, colorId))
      if err != nil { break L }
    }
  }

  ob.buffer = ob.buffer[:0]
  return err
}


// ========= Exporter
type Exporter struct {
  outBuf *outBuffer
  inChan imagedata.DataChannel
  *syncdata.SyncData
}

func New(outPath string, inChan imagedata.DataChannel, syncData *syncdata.SyncData) (*Exporter, error) {
  outBuf, err := newOutBuffer(outPath)
  if err != nil { return nil, err}

  return &Exporter{
    outBuf: outBuf,
    inChan: inChan,
    SyncData: syncData,
  }, nil
}

func (exp *Exporter) sendErr(err error) {
  if err == nil { return }
  exp.ErrChan <- exportError{err}
}

func (exp *Exporter) finish() {
  err := exp.outBuf.flush()
  if err != nil { exp.sendErr(err) }
  exp.Wg.Done()
}

func (exp *Exporter) Run() {
  go func() {
    defer exp.finish()
    exp.export()
  }()
}

func (exp *Exporter) export() {
  for data := range exp.inChan {
    exp.outBuf.write(data)
  }
}
