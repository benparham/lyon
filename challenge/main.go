package main

import (
  "log"
  "sync"
  "runtime"

  "github.com/docopt/docopt-go"
  "github.com/benparham/lyon/challenge/common/syncdata"
  "github.com/benparham/lyon/challenge/importer"
  "github.com/benparham/lyon/challenge/processor"
  "github.com/benparham/lyon/challenge/exporter"
)

const DEFAULT_IDPATH = "classification-images-b655c3f2e16467c2277e2102f2304f95.csv"
const DEFAULT_COLORPATH = "classification-colors-3268136d539e6e0084825fa4c5cb635a.csv"
const DEFAULT_OUTPATH = "result.txt"

func main() {

  usage := `Lyon + Post Challenge

Usage:
  challenge [--image-ids=<FILEPATH>] --store
  challenge [--image-ids=<FILEPATH>] [--colors=<FILEPATH>] [--output=<FILEPATH>] [--local]
  challenge -h | --help
  challenge --version

Options:
  --image-ids=<FILEPATH>  Path to file of image ids
  --colors=<FILEPATH>     Path to file of colors
  --output=<FILEPATH>     Path to file desired output file
  --store                 Download images to local directory
  --local                 Use images stored in local directory
  -h --help               Show this screen
  --version               Show version`

  args, _ := docopt.Parse(usage, nil, true, "Lyon + Post Challenge version 1.0", false)

  // Retrieve command line args
  idPath := args["--image-ids"]
  if idPath == nil { idPath = DEFAULT_IDPATH }

  // Download images to local files and return
  if args["--store"].(bool) {
    if err := importer.StoreImages(idPath.(string)); err != nil {
      log.Fatal(err)
    }
    return
  }

  colorPath := args["--colors"]
  if colorPath == nil { colorPath = DEFAULT_COLORPATH }

  outPath := args["--output"]
  if outPath == nil { outPath = DEFAULT_OUTPATH }

  local := args["--local"].(bool)

  // Run and log exceptions
  err := run(idPath.(string), colorPath.(string), outPath.(string), local)
  if err != nil { log.Fatal(err) }
}

func run(idPath string, colorPath string, outPath string, local bool) error {

  // Use available cores
  nCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPUs)

  // Construct sync data
  var wg sync.WaitGroup
  wg.Add(3)   // Importer, Processor and Exporter routines -> 3
  syncData := syncdata.New(make(chan error), &wg, make(chan struct{}, nCPUs))
  defer syncData.Cleanup()

  // Begin import routine
  imptr := importer.New(idPath, local, syncData)
  importChan := imptr.Run()

  // Begin processing routine
  prossr, err := processor.New(colorPath, importChan, syncData)
  if err != nil { return err }
  processorChan := prossr.Run()

  // Begin export routine
  exptr, err := exporter.New(outPath, processorChan, syncData)
  if err != nil { return err }
  exptr.Run()

  // Wait for spawned routines to complete
  doneChan := make(chan struct{})
  go func() {
    wg.Wait()
    close(doneChan)
  }()


  // Check for errors while waiting
L:
  for {
    select {
      case err := <- syncData.ErrChan:
        return err
      case <-doneChan:
        break L
    }
  }

  return nil
}
