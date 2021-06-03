package main

import (
	"flag"
	"github.com/xlab/portaudio-go/portaudio"
	"log"
	"os"
	"time"
)


const (
	framesPerPacket = 4096
	bitDepth        = 16
	audioChannels   = 2
	sampleFormat    = portaudio.PaInt16
)


func main() {

	var filename string
	// flags declaration using flag package
	flag.StringVar(&filename, "f", "", "Specify filename. No default")
	flag.Parse()  // after declaring flags we need to call it

	if err := portaudio.Initialize(); err != 0 {
		log.Fatalln("PortAudio init error:", err)
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	alacReader := NewAlacReader(f, bitDepth)
	defer alacReader.Close()

	if err := alacReader.Decode(); err != nil {
		log.Fatalln(err)
	}

	var stream *portaudio.Stream
	if err := portaudio.OpenDefaultStream(&stream, 0, audioChannels, sampleFormat, alacReader.SampleRate(),
		framesPerPacket, alacReader.StreamCallback, nil); err != 0 {
		log.Fatalln("PortAudio error:", err)
	}

	if err := portaudio.StartStream(stream); err != 0 {
		log.Fatalln("PortAudio error:", err)
	}

	time.Sleep(166 * time.Second)
}

