// Packet main for playground
package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/MarkKremer/microphone"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

func main() {

	testMic("micTest", 4)

	f, err := os.Open("micTest.wav")
	chk(err)

	streamer, format, err := wav.Decode(f)
	f.Close()
	chk(err)
	defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	//fmt.Println(format.SampleRate, format.NumChannels, format.Precision)

	//done := make(chan bool)
	speaker.Play(streamer)
	select {}
}

func testMic(filename string, duration int) {

	err := microphone.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer microphone.Terminate()

	stream, format, err := microphone.OpenDefaultStream(44100, 2)
	if err != nil {
		log.Fatal(err)
	}
	// Close the stream at the end if it hasn't already been
	// closed explicitly.
	defer stream.Close()

	if !strings.HasSuffix(filename, ".wav") {
		filename += ".wav"
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Stop the stream when the user tries to quit the program.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-sig:
			stream.Stop()
			stream.Close()

		case <-time.After(time.Duration(duration) * time.Second):
			stream.Stop()
			stream.Close()
		}
	}()

	stream.Start()

	// Encode the stream. This is a blocking operation because
	// wav.Encode will try to drain the stream. However, this
	// doesn't happen until stream.Close() is called.
	err = wav.Encode(f, stream, format)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func deleteFile(fileName string) error {
	// Check if the file exists
	if _, err := os.Stat(fileName); err == nil {
		// File exists, delete it
		err := os.Remove(fileName)
		if err != nil {
			return err
		}
	} else {
		// Some other error occurred
		return err
	}
	return nil
}
