package alac

/*
#cgo LDFLAGS: -L./codec -lalac
#include "alac.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Decoder struct {
	alacDecoder unsafe.Pointer
}

func NewDecoder(cookie []byte) (*Decoder, error) {
	cdecoder := C.new_decoder(unsafe.Pointer(&cookie[0]), (C.uint)(len(cookie)))
	if cdecoder == nil {
		return nil, errors.New("cookie init error")
	}
	return  &Decoder { alacDecoder: cdecoder}, nil
}

func (d *Decoder) Close() {
	C.delete_decoder(d.alacDecoder)
	d.alacDecoder = nil
}

func (d *Decoder) Decode(inBuffer []byte, outBuffer []byte, framesPerPacket uint32, channelsPerFrame uint32) ([]byte, error) {
	var outSize int
	samplesBuffer := (*C.uchar)(unsafe.Pointer(&outBuffer[0]))
	outNumSamples := (*C.uint)(unsafe.Pointer(&outSize))
	inputBuffer := (*C.uchar)(unsafe.Pointer(&inBuffer[0]))
	lenIn := (C.uint)(len(inBuffer))

	result := C.decode(d.alacDecoder, inputBuffer, lenIn, samplesBuffer, (C.uint)(framesPerPacket), (C.uint)(channelsPerFrame), outNumSamples )
	if result != 0 {
		return nil, errors.New("decode error")
	}
	return outBuffer[:outSize * 4], nil
}

func (d *Decoder) GetMaxFrameBytes() uint32 {
	info := C.get_decoder_config(d.alacDecoder)

	return uint32(info.maxFrameBytes)
}
