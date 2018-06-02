/*
Package squish provides bindings to the external libsquish library to be used for 
compression and decompression of DXT-encoded pixel data.

Important:
The original libSquish library is written in C++ and therefore not directly compatible with the CGo compiler tool.
The included patch file "libsquish-1.15-c_wrapper.patch" adds a simple C wrapper. Apply it before building the C++ library.
*/
package squish

/*
// For release: Use CGO_LDFLAGS environment variable to define -Llibdir and -llibname parameters
// Example: CGO_LDFLAGS=-L${SRCDIR} -lsquish -lgomp -lstdc++
#include "csquish.h"
*/
import "C"

import (
  "image"
  "unsafe"
)

var (
  // Compression type
  FLAGS_DXT1                  = int(C.ckDxt1)
  FLAGS_DXT3                  = int(C.ckDxt3)
  FLAGS_DXT5                  = int(C.ckDxt5)
  FLAGS_BC4                   = int(C.ckBc4)
  FLAGS_BC5                   = int(C.ckBc5)

  // Compression method. Quality ranges from "range fit" (lowest) to "iterative cluster fit" (highest).
  FLAGS_RANGE_FIT             = int(C.ckColourRangeFit)
  FLAGS_CLUSTER_FIT           = int(C.ckColourClusterFit)
  FLAGS_ITERATIVE_CLUSTER_FIT = int(C.ckColourIterativeClusterFit)

  // Additional flags
  FLAGS_WEIGHT_BY_ALPHA       = int(C.ckWeightColourByAlpha)
  FLAGS_SOURCE_BGRA           = int(C.ckSourceBGRA)

  // The default metric. Applies the same weight to all color components on compression.
  METRIC_UNIFORM    = []float32{ 1.0, 1.0, 1.0 }
  // A popular metric that improves percepted quality.
  METRIC_PERCEPTUAL = []float32{ 0.2126, 0.7152, 0.0722 }
)


// GetStorageRequirements computes the amount of compressed storage required.
//  param width    The width of the image.
//  param height   The height of the image.
//  param flags    Compression flags.
//  return         The required amount of compressed storage in bytes.
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// All other flags are ignored.
//
// Most DXT images will be a multiple of 4 in each dimension, but this function supports arbitrary size images by allowing the outer blocks to be 
// only partially used.
func GetStorageRequirements(width, height, flags int) int {
  var ret C.int = C.CGetStorageRequirements(C.int(width), C.int(height), C.int(flags))
  return int(ret)
}


// CompressImage compresses an image in memory.
//  param img      The source image.
//  param flags    Compression flags.
//  param metric   An optional perceptual metric.
//  return         Storage with the compressed output data.
func CompressImage(img image.Image, flags int, metric []float32) []byte {
  width, height := img.Bounds().Dx(), img.Bounds().Dy()
  stride := width * 4
  rgba := ImageToBytes(img)

  size := GetStorageRequirements(width, height, flags)
  blocks := make([]byte, size)
  return CompressBufferEx(rgba, width, height, stride, blocks, flags, metric)
}

// CompressBufferEx compresses an rgba pixel buffer in memory.
//  param rgba     The pixels of the source.
//  param width    The width of the source image.
//  param height   The height of the source image.
//  param pitch    The pitch of the source image (CompressImage2 only).
//  param blocks   Storage for the compressed output. Specify nil to auto-create.
//  param flags    Compression flags.
//  param metric   An optional perceptual metric.
//  return         The (updated) storage for the compressed output.
//
// The source pixels should be presented as a contiguous array of width*height rgba values, with each component as 1 byte each. In memory this should be: 
// { r1, g1, b1, a1, .... , rn, gn, bn, an } for n = width*height
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// When using DXT1 compression, 8 bytes of storage are required for each compressed DXT block. DXT3 and DXT5 compression require 16 bytes of storage per block.
//
// The flags parameter can also specify a preferred colour compressor to use when fitting the RGB components of the data. Possible colour compressors are: 
// FLAGS_CLUSTER_FIT (the default), FLAGS_RANGE_FIT (very fast, low quality) or FLAGS_ITERATIVE_CLUSTER_FIT (slowest, best quality).
//
// When using FLAGS_CLUSTER_FIT or FLAGS_ITERATIVE_CLUSTER_FIT, an additional flag can be specified to weight the importance of each pixel by its alpha value. 
// For images that are rendered using alpha blending, this can significantly increase the perceived quality.
//
// The metric parameter can be used to weight the relative importance of each colour channel, or pass NULL to use the default uniform weight of { 1.0f, 1.0f, 1.0f }. 
// This replaces the previous flag-based control that allowed either uniform or "perceptual" weights with the fixed values { 0.2126f, 0.7152f, 0.0722f }. 
// If non-NULL, the metric should point to a contiguous array of 3 floats.
//
// Internally this function calls CompressMasked() for each block, which allows for pixels outside the image to take arbitrary values. 
// The function GetStorageRequirements() can be called to compute the amount of memory to allocate for the compressed output.
//
// Note on compression quality: When compressing textures with libsquish it is recommended to apply a gamma-correction beforehand. This will reduce the blockiness 
// in dark areas. The level of necessary gamma-correction is platform dependent. For example, a gamma correction with gamma = 0.5 before compression 
// and gamma = 2.0 after decompression yields good results on the Windows platform but for other platforms like MacOS X a different gamma value may be more suitable.
func CompressBufferEx(rgba []byte, width, height, pitch int, blocks []byte, flags int, metric []float32) []byte {
  if blocks == nil {
    size := GetStorageRequirements(width, height, flags)
    blocks = make([]byte, size)
  }
  metricPtr := (*C.float)(unsafe.Pointer(nil))
  if metric != nil && len(metric) >= 3 {
    metricPtr = (*C.float)(unsafe.Pointer(&metric[0]))
  }
  C.CCompressImageEx((*C.uchar)(unsafe.Pointer(&rgba[0])), C.int(width), C.int(height), C.int(pitch), unsafe.Pointer(&blocks[0]), C.int(flags), metricPtr)
  return blocks
}

// CompressBuffer compresses an image in memory. It is identical to CompressBufferEx, except for the "pitch" parameter which is calculated from "width".
func CompressBuffer(rgba []byte, width, height int, blocks []byte, flags int, metric []float32) []byte {
  return CompressBufferEx(rgba, width, height, width * 4, blocks, flags, metric)
}


// DecompressImage decompresses an image in memory.
//  param width    The width of the source image.
//  param height   The height of the source image.
//  param blocks   The compressed DXT blocks as byte array.
//  param flags    Compression flags.
//  return         Image object containing the decompressed pixels. Returns nil on error.
func DecompressImage(width, height int, blocks []byte, flags int) image.Image {
  stride := width * 4
  buf := DecompressBufferEx(nil, width, height, stride, blocks, flags)

  var img image.Image = nil
  if buf != nil {
    nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
    ofsSrc, ofsDst := 0, 0
    for y := 0; y < height; y++ {
      copy(nrgba.Pix[ofsDst:ofsDst+nrgba.Stride], buf[ofsSrc:ofsSrc+stride])
      ofsSrc += stride
      ofsDst += nrgba.Stride
    }
    img = nrgba
  }
  return img
}

// DecompressBufferEx decompresses an image in memory.
//  param rgba     Storage for the decompressed pixels. Specify nil to auto-create.
//  param width    The width of the source image.
//  param height   The height of the source image.
//  param pitch    The pitch of the decompressed pixels.
//  param blocks   The compressed DXT blocks.
//  param flags    Compression flags.
//  return         The (updated) storage for the decompressed pixels.
//
// The decompressed pixels will be written as a contiguous array of width*height 16 rgba values, with each component as 1 byte each. In memory this is: 
// { r1, g1, b1, a1, .... , rn, gn, bn, an } for n = width*height
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// All other flags are ignored.
//
// Internally this function calls Decompress() for each block.
func DecompressBufferEx(rgba []byte, width, height, pitch int, blocks []byte, flags int) []byte {
  size := pitch * height
  if rgba == nil {
    rgba = make([]byte, size)
  }
  C.CDecompressImageEx((*C.uchar)(unsafe.Pointer(&rgba[0])), C.int(width), C.int(height), C.int(pitch), unsafe.Pointer(&blocks[0]), C.int(flags))
  return rgba
}

// DecompressBuffer decompresses an image in memory. It is identical to DecompressBufferEx, except for the "pitch" parameter which is calculated from "width".
func DecompressBuffer(rgba []byte, width, height int, blocks []byte, flags int) []byte {
  return DecompressBufferEx(rgba, width, height, width * 4, blocks, flags)
}


// CompressMasked compresses a 4x4 block of pixels.
//  param rgba     The rgba values of the 16 source pixels.
//  param mask     The valid pixel mask.
//  param block    Storage for the compressed DXT block. Specify nil to auto-create.
//  param flags    Compression flags.
//  param metric   An optional perceptual metric.
//  return         The (updated) storage for the compressed DXT block.
//
// The source pixels should be presented as a contiguous array of 16 rgba values, with each component as 1 byte each. In memory this should be: 
// { r1, g1, b1, a1, .... , r16, g16, b16, a16 }
//
// The mask parameter enables only certain pixels within the block. The lowest bit enables the first pixel and so on up to the 16th bit. 
// Bits beyond the 16th bit are ignored. Pixels that are not enabled are allowed to take arbitrary colours in the output block. 
// An example of how this can be used is in the CompressImage function to disable pixels outside the bounds of the image when the width or height 
// is not divisible by 4.
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// When using DXT1 compression, 8 bytes of storage are required for the compressed DXT block. DXT3 and DXT5 compression require 16 bytes of storage 
// per block.
//
// The flags parameter can also specify a preferred colour compressor to use when fitting the RGB components of the data. Possible colour compressors are: 
// FLAGS_CLUSTER_FIT (the default), FLAGS_RANGE_FIT (very fast, low quality) or FLAGS_ITERATIVE_CLUSTER_FIT (slowest, best quality).
//
// When using FLAGS_CLUSTER_FIT or FLAGS_ITERATIVE_CLUSTER_FIT, an additional flag can be specified to weight the importance of each pixel by its alpha value. 
// For images that are rendered using alpha blending, this can significantly increase the perceived quality.
//
// The metric parameter can be used to weight the relative importance of each colour channel, or pass NULL to use the default uniform weight of { 1.0f, 1.0f, 1.0f }. 
// This replaces the previous flag-based control that allowed either uniform or "perceptual" weights with the fixed values { 0.2126f, 0.7152f, 0.0722f }. 
// If non-NULL, the metric should point to a contiguous array of 3 floats.
func CompressMasked(rgba []byte, mask int, block []byte, flags int, metric []float32) []byte {
  if block == nil {
    size := GetStorageRequirements(4, 4, flags)
    block = make([]byte, size)
  }
  metricPtr := (*C.float)(unsafe.Pointer(nil))
  if metric != nil && len(metric) >= 3 {
    metricPtr = (*C.float)(unsafe.Pointer(&metric[0]))
  }
  C.CCompressMasked((*C.uchar)(unsafe.Pointer(&rgba[0])), C.int(mask), unsafe.Pointer(&block[0]), C.int(flags), metricPtr)
  return block
}


// Compress compresses a 4x4 block of pixels.
//  param rgba     The rgba values of the 16 source pixels.
//  param block    Storage for the compressed DXT block. Specify nil to auto-create.
//  param flags    Compression flags.
//  param metric   An optional perceptual metric.
//  return         The (updated) storage for the compressed DXT block.
//
// The source pixels should be presented as a contiguous array of 16 rgba values, with each component as 1 byte each. In memory this should be: 
// { r1, g1, b1, a1, .... , r16, g16, b16, a16 }
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// When using DXT1 compression, 8 bytes of storage are required for the compressed DXT block. DXT3 and DXT5 compression require 16 bytes of storage per block.
//
// The flags parameter can also specify a preferred colour compressor to use when fitting the RGB components of the data. Possible colour compressors are: 
// FLAGS_CLUSTER_FIT (the default), FLAGS_RANGE_FIT (very fast, low quality) or FLAGS_ITERATIVE_CLUSTER_FIT (slowest, best quality).
//
// When using FLAGS_CLUSTER_FIT or FLAGS_ITERATIVE_CLUSTER_FIT, an additional flag can be specified to weight the importance of each pixel by its alpha value. 
// For images that are rendered using alpha blending, this can significantly increase the perceived quality.
//
// The metric parameter can be used to weight the relative importance of each colour channel, or pass NULL to use the default uniform weight of { 1.0f, 1.0f, 1.0f }. 
// This replaces the previous flag-based control that allowed either uniform or "perceptual" weights with the fixed values { 0.2126f, 0.7152f, 0.0722f }. 
// If non-NULL, the metric should point to a contiguous array of 3 floats.
//
// This method is an inline that calls CompressMasked with a mask of 0xffff, provided for compatibility with older versions of squish.
func Compress(rgba []byte, block []byte, flags int, metric []float32) []byte {
  return CompressMasked(rgba, 0xffff, block, flags, metric)
}


// Decompress decompresses a 4x4 block of pixels.
//  param rgba     Storage for the 16 decompressed pixels. Specify nil to auto-create.
//  param block    The compressed DXT block.
//  param flags    Compression flags.
//  return         The (update) storage for the 16 decompressed pixels.
//
// The decompressed pixels will be written as a contiguous array of 16 rgba values, with each component as 1 byte each. In memory this is: 
// { r1, g1, b1, a1, .... , r16, g16, b16, a16 }
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// All other flags are ignored.
func Decompress(rgba []byte, block []byte, flags int) []byte {
  if rgba == nil {
    rgba = make([]byte, 64)
  }
  C.CDecompress((*C.uchar)(unsafe.Pointer(&rgba[0])), unsafe.Pointer(&block[0]), C.int(flags))
  return rgba
}


// ComputeMSEEx computes MSE of an compressed image in memory.
//  param rgba         The original image pixels.
//  param width        The width of the source image.
//  param height       The height of the source image.
//  param pitch        The pitch of the source image.
//  param dxt          The compressed dxt blocks
//  param flags        Compression flags.
//  return colourMSE   The MSE of the colour values.
//  return alphaMSE    The MSE of the alpha values.
//
// The colour MSE and alpha MSE are computed across all pixels. The colour MSE is averaged across all rgb values 
// (i.e. colourMSE = sum sum_k ||dxt.k - rgba.k||/3)
//
// The flags parameter should specify FLAGS_DXT1, FLAGS_DXT3, FLAGS_DXT5, FLAGS_BC4, or FLAGS_BC5 compression, however, DXT1 will be used by default if none is specified.
// All other flags are ignored.
//
// Internally this function calls Decompress() for each block.
func ComputeMSEEx(rgba []byte, width, height, pitch int, dxt []byte, flags int) (colorMSE, alphaMSE float64) {
  C.CComputeMSEEx((*C.uchar)(unsafe.Pointer(&rgba[0])), C.int(width), C.int(height), C.int(pitch), (*C.uchar)(unsafe.Pointer(&dxt[0])), C.int(flags), 
                  (*C.double)(unsafe.Pointer(&colorMSE)), (*C.double)(unsafe.Pointer(&alphaMSE)))
  return
}

// ComputeMSE computes MSE of an compressed image in memory. It is identical to ComputeMSEEx, except for the "pitch" parameter, which is calculated from "width".
func ComputeMSE(rgba []byte, width, height int, dxt []byte, flags int) (colorMSE, alphaMSE float64) {
  return ComputeMSEEx(rgba, width, height, width * 4, dxt, flags)
}
