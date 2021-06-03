#ifndef _ALAC_H
#define _ALAC_H

#include "codec/ALACBitUtilities.h"
#include "codec/ALACAudioTypes.h"

#ifdef __cplusplus
extern "C" {
#endif

void * new_decoder(void * inMagicCookie, uint32_t inMagicCookieSize);
int32_t	decode( void * decoder, uint8_t * bits, uint32_t len, uint8_t * sampleBuffer, uint32_t numSamples, uint32_t numChannels, uint32_t * outNumSamples );
ALACSpecificConfig* get_decoder_config(void * decoder);
void delete_decoder(void * decoder);

#ifdef __cplusplus
}
#endif

#endif