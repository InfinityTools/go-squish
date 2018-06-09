package squish

import (
  "image"
  "image/color"
)

// NRGBA converts a premultiplied color back to a normalized color with each component in range [0, 255].
func NRGBA(col color.Color) (r, g, b, a byte) {
  if nrgba, ok := col.(color.NRGBA); ok {
    r, g, b, a = nrgba.R, nrgba.G, nrgba.B, nrgba.A
  } else {
    pr, pg, pb, pa := col.RGBA()
    pa >>= 8
    if pa > 0 {
      pr >>= 8
      pr *= 0xff
      pr /= pa
      pg >>= 8
      pg *= 0xff
      pg /= pa
      pb >>= 8
      pb *= 0xff
      pb /= pa
    }
    r = byte(pr)
    g = byte(pg)
    b = byte(pb)
    a = byte(pa)
  }
  return
}


// ImageToBytes converts the image pixel data into a sequence of bytes where pixels are laid out as
// { r1, g1, b1, a1, ..., rn, gn, bn, an }.
func ImageToBytes(img image.Image) []byte {
  width, height := img.Bounds().Dx(), img.Bounds().Dy()
  stride := width * 4
  block := make([]byte, height*stride)

  nrgba, ok := img.(*image.NRGBA)
  if ok {
    // use NRGBA format directly
    ofsSrc, ofsDst := 0, 0
    for y := 0; y < height; y++ {
      copy(block[ofsDst:ofsDst+stride], nrgba.Pix[ofsSrc:ofsSrc+nrgba.Stride])
      ofsSrc += nrgba.Stride
      ofsDst += stride
    }
  } else {
    x0 := img.Bounds().Min.X
    y0 := img.Bounds().Min.Y
    ofsDst := 0
    for y := 0; y < height; y++ {
      for x := 0; x < width; x++ {
        block[ofsDst], block[ofsDst+1], block[ofsDst+2], block[ofsDst+3] = NRGBA(img.At(x+x0, y+y0))
        ofsDst += 4
      }
    }
  }
  return block
}
