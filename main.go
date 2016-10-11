package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cryptix/wav"
	"github.com/gordonklaus/portaudio"
	"io"
	"os"
	"os/signal"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: drumbox <file.wav>\n")
		os.Exit(1)
	}
	testInfo, err := os.Stat(os.Args[1])
	checkErr(err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	testWav, err := os.Open(os.Args[1])
	checkErr(err)

	wavReader, err := wav.NewReader(testWav, testInfo.Size())
	fmt.Println(wavReader.String())
	checkErr(err)

	audioData := []uint16{}

	// Read in complete audio file
	for {
		samp, err := wavReader.ReadRawSample()
		if err == io.EOF {
			break
		}
		checkErr(err)

		var sampVal uint16
		sampBuf := bytes.NewBuffer(samp)
		binary.Read(sampBuf, binary.LittleEndian, &sampVal)
		audioData = append(audioData, sampVal)
	}

	numSamples := wavReader.GetSampleCount()

	portaudio.Initialize()
	defer portaudio.Terminate()
	paBuffer := make([]uint16, 8192)
	stream, err := portaudio.OpenDefaultStream(
		0,
		1,
		44100,
		len(paBuffer),
		&paBuffer,
	)
	checkErr(err)

	audioDataBuffer := bytes.NewBuffer(audioData)

	for remaining := int(numSamples); remaining > 0; remaining -= len(paBuffer) {
		if len(paBuffer) > remaining {
			paBuffer = paBuffer[:remaining]
		}
		err := binary.Read(audioDataBuffer, binary.LittleEndian, paBuffer)
		if err == io.EOF {
			break
		}
		checkErr(err)
		checkErr(stream.Write())

		select {
		case <-sig:
			return
		default:
		}
		stream.Write()
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
