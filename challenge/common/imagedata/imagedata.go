package imagedata

/*
  Struct defining image object
*/

import (
  "fmt"
  "image"
)

type ImageData struct {
  Id string           // Image id
  Data *image.Image    // Image data
  ColorIds []string   // Ids of computed color buckets
}

// Debugging methods
func (d ImageData) String() string {
  return fmt.Sprintf("%s - ColorIds: %s", d.Id, d.ColorIds)
}

type DataChannel chan *ImageData
