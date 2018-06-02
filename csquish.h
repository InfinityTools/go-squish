#ifndef CSQUISH_H
#define CSQUISH_H

/*
 * Hides the C++ function calls of the squish library behind C compatible function definitions.
 */
#ifdef __cplusplus
extern "C" {
#endif

typedef unsigned char u8;

extern const int ckDxt1;
extern const int ckDxt3;
extern const int ckDxt5;
extern const int ckBc4;
extern const int ckBc5;
extern const int ckColourClusterFit;
extern const int ckColourRangeFit;
extern const int ckWeightColourByAlpha;
extern const int ckColourIterativeClusterFit;
extern const int ckSourceBGRA;

void CCompressMasked( u8 const* rgba, int mask, void* block, int flags, float* metric );
inline void CCompress( u8 const* rgba, void* block, int flags, float* metric )
{
    CCompressMasked( rgba, 0xffff, block, flags, metric );
}

void CDecompress( u8* rgba, void const* block, int flags );

int CGetStorageRequirements( int width, int height, int flags );

void CCompressImageEx( u8 const* rgba, int width, int height, int pitch, void* blocks, int flags, float* metric );
inline void CCompressImage( u8 const* rgba, int width, int height, void* blocks, int flags, float* metric )
{
    CCompressImageEx( rgba, width, height, width * 4, blocks, flags, metric );
}

void CDecompressImageEx( u8* rgba, int width, int height, int pitch, void const* blocks, int flags );
inline void CDecompressImage( u8* rgba, int width, int height, void const* blocks, int flags )
{
    CDecompressImageEx( rgba, width, height, width * 4, blocks, flags );
}

void CComputeMSEEx( u8 const *rgba, int width, int height, int pitch, u8 const *dxt, int flags, double *colourMSE, double *alphaMSE );
inline void CComputeMSE( u8 const *rgba, int width, int height, u8 const *dxt, int flags, double *colourMSE, double *alphaMSE )
{
  CComputeMSEEx( rgba, width, height, width * 4, dxt, flags, colourMSE, alphaMSE );
}


#ifdef __cplusplus
}
#endif

#endif // ndef CSQUISH_H
