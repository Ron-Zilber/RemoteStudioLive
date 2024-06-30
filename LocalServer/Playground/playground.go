// Packet main for playground
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	"layeh.com/gopus"
	//"github.com/hraban/opus"
	//"github.com/hraban/opus"
	//"gopkg.in/hraban/opus.v2"
)

const (
	SampleRate = 48000 // SampleRate is the number of bits used to represent a full second of audio sampling
	Channels   = 2     // Channels - 1 for mono; 2 for stereo
	FrameSize  = 960   // FrameSize of 960 gives 20 ms (for 48kHz sampling) which is the Opus recommendation

	BufferSize = FrameSize * Channels // BufferSize let the buffer hold multiple frames
)

func main() {
	frameSize := 480
	durationSeconds := 3
	portaudio.Initialize()
	defer portaudio.Terminate()
	audioBufferSize := frameSize * Channels
	_ = audioBufferSize
	in := make([]int16, audioBufferSize)
	stream, err := portaudio.OpenDefaultStream(Channels, 0, SampleRate, len(in)/4, in)
	CheckError(err)
	defer stream.Close()

	encoder, err := gopus.NewEncoder(SampleRate, Channels, gopus.Audio)
	CheckError(err)
	tInit := time.Now().UnixMicro()
	CheckError(stream.Start())

	packetsCounter := 0
	fmt.Println("Record start")
	for {
		tRecordFrame := time.Now().UnixMicro()

		CheckError(stream.Read())                             //* Read filling the buffer by recording samples until the buffer is full
		data, err := encoder.Encode(in, frameSize, audioBufferSize) //* Encode PCM to Opus
		CheckError(err)
		_ = data

		tProcessing := time.Now().UnixMicro() - tRecordFrame

		fmt.Println("Processing time:", tProcessing)
		packetsCounter++

		if time.Now().UnixMicro()-tInit > int64(durationSeconds)*1000000 {
			fmt.Println("Record end")
			CheckError(stream.Stop())
			return
		}

	}
}

func printFile(fileName string) {
	f, err := os.Open(fileName)
	CheckError(err)

	_, err = io.Copy(os.Stdout, f)
	CheckError(err)
	CheckError(f.Close())
}

func deleteFile(fileName string) error {
	// Check if the file exists
	if _, err := os.Stat(fileName); err == nil {
		// File exists, delete it

		err := os.Remove(fileName)
		if err != nil {
			return err
		}
	} else if !strings.HasSuffix(err.Error(), "no such file or directory") {
		// Some other error occurred
		return err
	}
	return nil
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}
