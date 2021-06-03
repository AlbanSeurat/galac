#include "alac.h"
#include "codec/ALACDecoder.h"
#include "codec/ALACBitUtilities.h"
#include <stdio.h>

void * new_decoder(void * inMagicCookie, uint32_t inMagicCookieSize) {
    ALACDecoder * pDecoder = new ALACDecoder;
    int ret = pDecoder->Init(inMagicCookie, inMagicCookieSize);
    if (ret != 0) {
        delete pDecoder;
        return (void*)0;
    }
    return pDecoder;
}
int32_t	decode( void * decoder, uint8_t * bits, uint32_t len, uint8_t * sampleBuffer, uint32_t numSamples, uint32_t numChannels, uint32_t * outNumSamples ) {

    struct BitBuffer theInputBuffer;
    BitBufferInit(&theInputBuffer, bits, len);
    return ((ALACDecoder *)decoder)->Decode(&theInputBuffer, sampleBuffer, numSamples, numChannels, outNumSamples);
}

ALACSpecificConfig* get_decoder_config(void * decoder) {
    return &((ALACDecoder *)decoder)->mConfig;
}

void delete_decoder(void * decoder) {
    delete ((ALACDecoder *)decoder);
}

