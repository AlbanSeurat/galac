package main

import (
	"bytes"
	"errors"
	"fmt"
	"galac/alac"
	"github.com/jfbus/mp4"
	"github.com/xlab/portaudio-go/portaudio"
	"io"
	"log"
	"unsafe"
)

var (
	AlacErr = errors.New("ALAC header corrupt")
	AlacTag = []byte("alac")
)


type uint32Row struct {
	N     uint32
	Value uint32
}

type uint32Rows []uint32Row

// SearchFirst is used to get number of audio packets per chunk N
// provided the following list of pairs:
//
// #0 : 5 packets per chunk starting @chunk #1
// #1 : 2 packets per chunk starting @chunk #131
func (s uint32Rows) SearchFirst(N uint32) (Value uint32) {
	if len(s) == 0 {
		return 0
	}
	for i := range s {
		if s[i].N < N {
			continue
		} else if s[i].N == N {
			return s[i].Value
		}
		return s[i-1].Value
	}
	return s[len(s)-1].Value
}

// SearchLast is used to get the duration of audio frame N
// provided the following list of pairs:
//
// #0 : 651 frames with duration 4096 units
// #1 : 1 frames with duration 672 units
func (s uint32Rows) SearchLast(N uint32) (Value uint32) {
	if len(s) == 0 {
		return 0
	}
	for i := range s {
		if n := s[i].N; N < n {
			return s[i].Value
		} else if N = N - n; N >= 0 {
			continue
		}
	}
	return s[len(s)-1].Value
}

type AlacReader struct {
	buf   []byte
	debuf []byte
	in    io.ReadSeeker
	a     *alac.Decoder

	frameN uint32
	chunkN uint32

	bytesPerSample    uint32
	sampleRate        uint32
	framesTotal       uint32
	unitsTotal        uint32
	maxSamplesInFrame uint32
	frameDurations    uint32Rows
	packetsPerChunk   uint32Rows
	chunkOffsets      []uint32
	packetSizes       []uint32
	packetUniformSize uint32

}


func NewAlacReader(in io.ReadSeeker, bitDepth int) *AlacReader {
	return &AlacReader{
		in: in,
		bytesPerSample: uint32(bitDepth / 8),
	}
}

func (a *AlacReader) Duration() float32 {
	return float32(a.unitsTotal) / float32(a.sampleRate)
}

func (a *AlacReader) chunkOffset(n uint32) int64 {
	return int64(a.chunkOffsets[n])
}

// sampleSize gets the size of n-th sample packet in chunk.
func (a *AlacReader) packetSize(n uint32) uint32 {
	if len(a.packetSizes) == 0 {
		return a.packetUniformSize
	}
	return a.packetSizes[n]
}

func (a *AlacReader) SampleRate() float64 {
	return float64(a.sampleRate)
}



func (a *AlacReader) Decode() error {
	v, err := mp4.Decode(a.in)
	if err != nil {
		return err
	}
	if v.Moov == nil || len(v.Moov.Trak) == 0 {
		return errors.New("no track found")
	}
	var mdia *mp4.MdiaBox
	for i := range v.Moov.Trak {
		m := v.Moov.Trak[i].Mdia
		if m != nil && m.Hdlr != nil && m.Hdlr.HandlerType == "soun" {
			mdia = m
		}
	}
	if mdia == nil {
		return errors.New("no audio track found")
	}
	a.sampleRate = mdia.Mdhd.Timescale
	a.unitsTotal = mdia.Mdhd.Duration

	table := mdia.Minf.Stbl

	size := table.Stsd.Size()
	buf := bytes.NewBuffer(make([]byte, 0, size))
	table.Stsd.Encode(buf)
	if len(buf.Next(0x34)) < 0x34 {
		return AlacErr
	} else if v := buf.Next(4); len(v) < 4 {
		return AlacErr
	} else {
		size := int(uint(v[0])<<24 | uint(v[1])<<16 | uint(v[2])<<8 | uint(v[3]))
		if cookie := buf.Next(size - 4); len(cookie) < size-4 {
			return AlacErr
		} else {
			log.Printf("ALAC header: %02X", cookie[8:])
			a.a , err = alac.NewDecoder(cookie[8:])
			if err != nil {
				return AlacErr
			}
			a.maxSamplesInFrame = a.a.GetMaxFrameBytes()
			// audio packet buffers
			a.buf = make([]byte, a.maxSamplesInFrame*a.bytesPerSample)
			a.debuf = make([]byte, a.maxSamplesInFrame*a.bytesPerSample)
		}
	}

	for i, count := range table.Stts.SampleCount {
		a.frameDurations = append(a.frameDurations, uint32Row{
			N:     count,
			Value: table.Stts.SampleTimeDelta[i],
		})
	}
	for i, first := range table.Stsc.FirstChunk {
		// audio frames per chunk
		a.packetsPerChunk = append(a.packetsPerChunk, uint32Row{
			N:     first - 1,
			Value: table.Stsc.SamplesPerChunk[i],
		})
	}

	a.framesTotal = table.Stsz.SampleNumber
	if size := table.Stsz.SampleUniformSize; size > 0 {
		a.packetUniformSize = size
	} else {
		a.packetSizes = table.Stsz.SampleSize
	}
	a.chunkOffsets = table.Stco.ChunkOffset

	if _, err := a.in.Seek(a.chunkOffset(0), 0); err != nil {
		return errors.New("cannot read audio data from file")
	}
	fmt.Printf("Audio duration: %.3fs\n", a.Duration())
	return nil
}

func (a *AlacReader) ReadFrame(framesPerPacket uint32, audioChannels uint32) ([]byte, error) {
	if a.frameN >= a.framesTotal {
		return nil, AlacErr
	}
	packetSize := a.packetSize(a.frameN)
	if _, err := a.in.Read(a.buf[:packetSize]); err != nil {
		log.Println("[warn]:", err)
		return nil, AlacErr
	}
	if _, err := a.a.Decode(a.buf[:packetSize], a.debuf, framesPerPacket, audioChannels) ; err != nil {
		log.Println("[warn]:", err)
		return nil, AlacErr
	}
	a.frameN++
	return nil, nil
}


func (a *AlacReader) StreamCallback(_ unsafe.Pointer, output unsafe.Pointer, sampleCount uint,
	_ *portaudio.StreamCallbackTimeInfo, _ portaudio.StreamCallbackFlags, _ unsafe.Pointer) int32 {

	const (
		statusContinue = int32(portaudio.PaContinue)
		statusComplete = int32(portaudio.PaComplete)
		statusAbort    = int32(portaudio.PaAbort)
	)

	if a.frameN >= a.framesTotal {
		return statusComplete
	}
	// if a.advanceChunk() != nil {
	// 	return statusComplete
	// }
	packetSize := a.packetSize(a.frameN)
	if _, err := a.in.Read(a.buf[:packetSize]); err != nil {
		log.Println("[warn]:", err)
		return statusAbort
	}

	result , err := a.a.Decode(a.buf[:packetSize], a.debuf, framesPerPacket, audioChannels)
	if err != nil {
		log.Println("[warn]:", err)
		return statusAbort
	}
	// sampleCount a.k.a samples in the frame, a frame usually has
	// 4096 samples per channel, so we process (int16|int16) at a time for L|R.
	out := (*(*[1 << 24]int16)(output))[:sampleCount*audioChannels]
	for i := 0; i < len(out); i++ {
		// 2 channel 16-bit stereo
		out[i] = int16(result[2*i]) | int16(result[2*i+1])<<8
	}
	a.frameN++
	// a.currentFrameInChunk++
	return statusContinue
}


func (a *AlacReader) Close() {
	a.a.Close()
}