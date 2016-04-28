package syncdata

/*
  Struct containing common synchronization items
*/

import (
  "sync"
)

type SyncData struct {
  ErrChan chan error
  Wg *sync.WaitGroup
  ResourceSem chan struct{}
}

func New(errChan chan error, wg *sync.WaitGroup, resourceSem chan struct{}) *SyncData {
  return &SyncData{
    ErrChan: errChan,
    Wg: wg,
    ResourceSem: resourceSem,
  }
}

func (sd *SyncData) Cleanup() {
  close(sd.ErrChan)
  close(sd.ResourceSem)
}
