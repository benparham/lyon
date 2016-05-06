package processor

import (
  "encoding/csv"
  "errors"
  "io"
  "os"

  converter "github.com/benparham/lyon/challenge/colorconverter"
  chromath "github.com/jkl1337/go-chromath"
)

const CSV_RECORD_LENGTH = 3
const MULTI_CODE = "xxxxxx"


// ======== Color bucket
type colorBucket struct {
  colorId string
  colorName string
  labColor chromath.Lab
}


type loadResult struct {
  buckets []colorBucket
  multiId string
  whiteId string
  err error
}

// Load color data from file and convert to CIELab
func loadColorBuckets(colorPath string) ([]colorBucket, string, string, error) {
  buckets := make([]colorBucket, 0)
  multiId := ""
  whiteId := ""

  colorFile, err := os.Open(colorPath)
  if err != nil { return buckets, multiId, whiteId, err}
  defer colorFile.Close()

  rdr := csv.NewReader(colorFile)
  idx := 0
  for {
    record, err := rdr.Read()
    if err == io.EOF { break }
    if err != nil { return buckets, multiId, whiteId, err }

    if len(record) != CSV_RECORD_LENGTH {
      return buckets, multiId, whiteId, errors.New("Malformed color csv file")
    }

    if record[2] == MULTI_CODE {
      multiId = record[0]
      continue
    }
    if record[1] == "White" || record[1] == "white" {
      whiteId = record[0]
    }

    lab, err := converter.HexStringToLab(record[2])
    if err != nil { return buckets, multiId, whiteId, err }

    buckets = append(buckets, colorBucket{
      colorId: record[0],
      colorName: record[1],
      labColor: lab,
    })

    idx++
  }

  if multiId == "" { return buckets, multiId, whiteId, errors.New("No multival given")}
  return buckets, multiId, whiteId, nil
}
