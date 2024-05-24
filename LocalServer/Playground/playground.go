// Packet main for playground
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func main() {

	Mp3File := "outfile.mp3"
	CheckError(deleteFile(Mp3File))

	rawData, err := recordRawToBytes(5)
	CheckError(err)
	encodeBytesToMp3(rawData, Mp3File)

	printFile(Mp3File)
}

func encodeToMp3(rawFile string, Mp3File string) {
	fRaw := rawFile
	fMp3 := Mp3File

	err := ffmpeg.Input(fRaw, ffmpeg.KwArgs{
		"f":  "s16le",
		"ar": "44100",
		"ac": "1",
	}).Output(fMp3, ffmpeg.KwArgs{
		"b:a": "192k",
	}).Run()
	if err != nil {
		fmt.Printf("Error encoding to MP3: %v\n", err)
		return
	}
}

func encodeBytesToMp3(rawData []byte, Mp3File string) {
	rawFile, err := os.CreateTemp("", "temp")
	CheckError(err)

	_, err = rawFile.Write(rawData)
	CheckError(err)
	CheckError(rawFile.Close())
	defer os.Remove(rawFile.Name())

	err = ffmpeg.Input(rawFile.Name(), ffmpeg.KwArgs{
		"f":  "s16le",
		"ar": "44100",
		"ac": "1",
	}).Output(Mp3File, ffmpeg.KwArgs{
		"b:a": "192k",
	}).Run()
	if err != nil {
		fmt.Printf("Error encoding to MP3: %v\n", err)
		return
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
	} else {
		// Some other error occurred
		return err
	}
	return nil
}

func recordRaw(fileName string, duration int64) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	f, err := os.Create(fileName)
	CheckError(err)

	defer func() {
		CheckError(f.Close())
	}()

	portaudio.Initialize()
	//time.Sleep(1)
	defer portaudio.Terminate()
	in := make([]int16, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	CheckError(err)
	defer stream.Close()

	CheckError(stream.Start())
	tStart := time.Now().Unix()

loop:
	for {
		CheckError(stream.Read())
		CheckError(binary.Write(f, binary.LittleEndian, in))
		select {
		case <-sig:
			break loop

		default:
			if time.Now().Unix()-tStart > duration {
				//CheckError(stream.Stop())
				break loop

			}
		}
	}
	CheckError(stream.Stop())
}

func recordRawToBytes(duration int64) ([]byte, error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	var buf bytes.Buffer
	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int16, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return nil, err
	}
	tStart := time.Now().Unix()

loop:
	for {
		if err := stream.Read(); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.LittleEndian, in); err != nil {
			return nil, err
		}
		select {
		case <-sig:
			break loop

		default:
			if time.Now().Unix()-tStart > duration {
				break loop
			}
		}
	}
	if err := stream.Stop(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func record(fileName string, duration int64) {

	fmt.Println("Recording.  Press Ctrl-C to stop.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	if !strings.HasSuffix(fileName, ".aiff") {
		fileName += ".aiff"
	}
	f, err := os.Create(fileName)
	CheckError(err)

	// form chunk
	_, err = f.WriteString("FORM")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //total bytes
	_, err = f.WriteString("AIFF")
	CheckError(err)

	// common chunk
	_, err = f.WriteString("COMM")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(18)))           //size
	CheckError(binary.Write(f, binary.BigEndian, int16(1)))            //channels
	CheckError(binary.Write(f, binary.BigEndian, int32(0)))            //number of samples
	CheckError(binary.Write(f, binary.BigEndian, int16(32)))           //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	CheckError(err)

	// sound chunk
	_, err = f.WriteString("SSND")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //size
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //offset
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //block
	nSamples := 0
	defer func() {
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		_, err = f.Seek(4, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(totalBytes)))
		_, err = f.Seek(22, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(nSamples)))
		_, err = f.Seek(42, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(4*nSamples+8)))
		CheckError(f.Close())
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	CheckError(err)
	defer stream.Close()

	CheckError(stream.Start())

	tStart := time.Now().Unix()
	for {
		CheckError(stream.Read())
		CheckError(binary.Write(f, binary.BigEndian, in))
		nSamples += len(in)
		select {
		case <-sig:
			CheckError(stream.Stop())
			return

		default:
			if time.Now().Unix()-tStart > duration {
				CheckError(stream.Stop())
				return
			}
		}

	}
	fmt.Println("Stopping the stream")
	CheckError(stream.Stop())
}
func (id ID) String() string {
	return string(id[:])
}
func readChunk(r readerAtSeeker) (id ID, data *io.SectionReader, err error) {
	_, err = r.Read(id[:])
	if err != nil {
		return
	}
	var n int32
	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}
	off, _ := r.Seek(0, 1)
	data = io.NewSectionReader(r, off, int64(n))
	_, err = r.Seek(int64(n), 1)
	return
}
func play(fileName string) {
	fmt.Println("Playing.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	f, err := os.Open(fileName)
	CheckError(err)
	defer f.Close()

	id, data, err := readChunk(f)
	CheckError(err)
	if id.String() != "FORM" {
		fmt.Println("bad file format")
		return
	}
	_, err = data.Read(id[:])
	CheckError(err)
	if id.String() != "AIFF" {
		fmt.Println("bad file format")
		return
	}
	var c commonChunk
	var audio io.Reader
	for {
		id, chunk, err := readChunk(data)
		if err == io.EOF {
			break
		}
		CheckError(err)
		switch id.String() {
		case "COMM":
			CheckError(binary.Read(chunk, binary.BigEndian, &c))
		case "SSND":
			chunk.Seek(8, 1) //ignore offset and block
			audio = chunk
		default:
			fmt.Printf("ignoring unknown chunk '%s'\n", id)
		}
	}

	//assume 44100 sample rate, mono, 32 bit

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int32, 8192)
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), &out)
	CheckError(err)
	defer stream.Close()

	CheckError(stream.Start())
	defer stream.Stop()
	for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
		if len(out) > remaining {
			out = out[:remaining]
		}
		err := binary.Read(audio, binary.BigEndian, out)
		if err == io.EOF {
			break
		}
		CheckError(err)
		CheckError(stream.Write())
		select {
		case <-sig:
			return
		default:
		}
	}
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type commonChunk struct {
	NumChans      int16
	NumSamples    int32
	BitsPerSample int16
	SampleRate    [10]byte
}

type ID [4]byte
