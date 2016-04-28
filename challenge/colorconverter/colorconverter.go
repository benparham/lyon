package colorconverter

/*
  Tools for converting to CIELab and comparing those Lab colors
*/

import (
  "encoding/hex"
  "image/color"

  "github.com/jkl1337/go-chromath/deltae"
  chromath "github.com/jkl1337/go-chromath"
)

// ===== chromath Transformers
var rgbToXyzTransformer = chromath.NewRGBTransformer(
  &chromath.SpaceSRGB,
  &chromath.AdaptationBradford,
  &chromath.IlluminantRefD50,
  &chromath.Scaler8bClamping,
  1.0,
  nil,
)

var labToXyzTransformer = chromath.NewLabTransformer(&chromath.IlluminantRefD50)



// ======== Converters

// unexported
func hexStringToRgb(hexStr string) (chromath.RGB, error) {
  var rgb chromath.RGB

  bytes, err := hex.DecodeString(hexStr)
  if err != nil { return rgb, err }
  rgb = chromath.RGB{float64(bytes[0]), float64(bytes[1]), float64(bytes[2])}

  return rgb, nil
}

func rgbaToRgb(rgba color.Color) chromath.RGB {
  r, g, b, _ := rgba.RGBA()
  return chromath.RGB{float64(r>>8), float64(g>>8), float64(b>>8)}
}

func rgbToLab(rgb chromath.RGB) chromath.Lab {
  return labToXyzTransformer.Invert(rgbToXyzTransformer.Convert(rgb))
}


// exported
func HexStringToLab(hexStr string) (chromath.Lab, error) {
  lab := chromath.Lab{0,0,0}

  rgb, err := hexStringToRgb(hexStr)
  if err != nil { return lab, err }
  return rgbToLab(rgb), nil
}

func RgbaToLab(rgba color.Color) chromath.Lab {
  return rgbToLab(rgbaToRgb(rgba))
}


// Distance Calculation
func CalcDist(c1 chromath.Lab, c2 chromath.Lab) float64 {
  return deltae.CIE2000(c1, c2, &deltae.KLChDefault)
}
