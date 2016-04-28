package importer

import (
  "fmt"
  "io"
  "os"
  "bufio"
  "time"
  "log"
  "image/jpeg"
  "io/ioutil"
  "net/http"
  "path/filepath"

  "runtime"

  "github.com/lyon/challenge/common/imagedata"
  "github.com/lyon/challenge/common/syncdata"
)

// ========== Constants
const STORE_DIR = "images"
const STORE_PERM = 0755

const URL_PREFIX = "http://lyonandpost.imgix.net/product-"
const URL_SUFFIX = ".jpg?h=600"

const RETRY_CUTOFF = 3
const RETRY_DELAY = 1000 * time.Millisecond

const IMPORT_BUFFER = 200 // Number of images to store in memory


// ========== Custom Error
type importError struct { err error }
func (ie importError) Error() string {
  return fmt.Sprintf("[Importer] %s", ie.err.Error())
}


// ========== Utility
func imageUrl(id string) string {
  return fmt.Sprintf("%s%s%s", URL_PREFIX, id, URL_SUFFIX)
}

func imagePath(id string) string {
  return fmt.Sprintf("%s/%s.jpg", STORE_DIR, id)
}


// ========== Importer
type Importer struct {
  idPath string
  local bool
  *syncdata.SyncData
  outChan imagedata.DataChannel
}

func New(idPath string, local bool, syncData *syncdata.SyncData) *Importer {
  return &Importer{
    idPath: idPath,
    local: local,
    SyncData: syncData,
    outChan: make(imagedata.DataChannel, IMPORT_BUFFER),
  }
}

func (imp *Importer) send(data *imagedata.ImageData) {
  imp.outChan <- data
}

func (imp *Importer) sendErr(err error) {
  if err == nil { return }

  var stack [4096]byte
  runtime.Stack(stack[:], false)
  log.Printf("%q\n%s\n", err, stack[:])
  // return err

  imp.ErrChan <- importError{ err }
}

func (imp *Importer) finish() {
  close(imp.outChan)
  imp.Wg.Done()
}

func (imp *Importer) Run() imagedata.DataChannel {
  if (imp.local) {
    go func() {
      defer imp.finish()
      imp.fetchLocalImages()
    }()
  } else {
    go func() {
      defer imp.finish()
      imp.fetchImages()
    }()
  }

  return imp.outChan
}


// Tells downloadImages() what to do with each response
type responseProcessor func(id string, bodyRdr io.Reader) error

// Download images from api
func downloadImages(idPath string, respProc responseProcessor) (err error) {
  idFile, err := os.Open(idPath)
  if err != nil { return }
  defer idFile.Close()

  err = os.MkdirAll(STORE_DIR, STORE_PERM)
  if err != nil { return }

  scanner := bufio.NewScanner(idFile)
  for scanner.Scan() {
    id := scanner.Text()
    var resp *http.Response

    attempts := 0
    for {
      resp, err = http.Get(imageUrl(id))
      if err != nil {
        log.Print(err)
        time.Sleep(RETRY_DELAY)
        attempts++

        if attempts < RETRY_CUTOFF {
          continue
        } else {
          return err
        }
      }
      defer resp.Body.Close()
      break
    }

    err = respProc(id, resp.Body)
    if err != nil { return err }

    // break
  }

  return
}

// Store downloaded images in local directory
func StoreImages(idPath string) error {
  respProc := func(id string, bodyRdr io.Reader) error {
    body, err := ioutil.ReadAll(bodyRdr)
    if err != nil { return importError{err} }
    return ioutil.WriteFile(imagePath(id), body, STORE_PERM)
  }

  return downloadImages(idPath, respProc)
}

// Pass downloaded images over channel
func (imp *Importer) fetchImages() {
  err := downloadImages(
    imp.idPath,
    func(id string, bodyRdr io.Reader) error {
      img, err := jpeg.Decode(bodyRdr)
      if err == nil {
        imp.send(&imagedata.ImageData{Id: id, Data: &img})
      } else {
        log.Printf("Error decoding image %s to jpeg\n", id)
      }
      return nil
    },
  )

  if err != nil { imp.sendErr(err) }
}

// Pass images from local directory over channel
func (imp *Importer) fetchLocalImages() {
  items, err := ioutil.ReadDir(STORE_DIR)
  if err != nil {
    imp.sendErr(err)
    return
  }

  for _, info := range items {
    if info.IsDir() { continue }
    filename := info.Name()

    file, err := os.Open(fmt.Sprintf("%s/%s", STORE_DIR, filename))
    if err != nil {
      imp.sendErr(err)
      return
    }

    img, err := jpeg.Decode(file)
    if err != nil {
      log.Printf("Error decoding image file %s to jpeg\n", filename)
      continue
    }

    ext := filepath.Ext(filename)
    id := filename[:len(filename)-len(ext)]

    imp.send(&imagedata.ImageData{Id: id, Data: &img})
  }
}
