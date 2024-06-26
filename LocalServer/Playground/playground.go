// Packet main for playground
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
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

	// Clears the screen
	fmt.Print("Record starts in:\n3")
	//fmt.Print("3")
	time.Sleep(time.Second)
	fmt.Print("\r2")
	time.Sleep(time.Second)
	fmt.Print("\r1")
	time.Sleep(time.Second)
	fmt.Println("\rStart playing")
}

func recordAndSend(destinationChannel chan []byte, durationMseconds int, waitGroup *sync.WaitGroup) {
	defer func() {
		waitGroup.Done()
		fmt.Println("recordAndSend Done")
		close(destinationChannel)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()

	in := make([]int16, BufferSize) // Each Buffer records 10 milliseconds
	stream, err := portaudio.OpenDefaultStream(Channels, 0, SampleRate, len(in), in)
	CheckError(err)
	defer stream.Close()

	encoder, err := gopus.NewEncoder(SampleRate, Channels, gopus.Audio)
	CheckError(err)

	tStart := time.Now().UnixMilli()
	CheckError(stream.Start())

	for {
		CheckError(stream.Read()) //* Read filling the buffer by recording samples until the buffer is full

		data, err := encoder.Encode(in, FrameSize, BufferSize)
		CheckError(err)
		//fmt.Println("dataSize: ", len(data), "bufferSize: ", len(in))
		destinationChannel <- data
		select {
		case <-sig:
			CheckError(stream.Stop())
			return
		default:
			if time.Now().UnixMilli()-tStart > int64(durationMseconds) {
				CheckError(stream.Stop())
				return
			}
		}
	}
}

func play(channel chan []byte, waitGroup *sync.WaitGroup) {
	defer func() {
		waitGroup.Done()
		fmt.Println("play Done")
	}()

	decoder, err := gopus.NewDecoder(SampleRate, Channels)
	CheckError(err)

	speaker.Init(beep.SampleRate(SampleRate), BufferSize)
	var buffer [][2]float64
	streamer := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if len(buffer) == 0 {
			chunk, ok := <-channel
			if !ok {
				// The channel has been closed
				fmt.Println("Channel Closed")
				return 0, false
			}
			pcm, err := decoder.Decode(chunk, FrameSize, false)
			if err != nil {
				fmt.Println("Error decoding chunk:", err)
				return 0, false
			}

			for i := 0; i < len(pcm); i += 2 {
				buffer = append(buffer, [2]float64{
					float64(pcm[i]) / 32768.0,
					float64(pcm[i+1]) / 32768.0,
				})
			}
		}

		for i := range samples {
			if len(buffer) == 0 {
				return i, true
			}
			samples[i] = buffer[0]
			buffer = buffer[1:]
		}
		return len(samples), true
	})

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
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
