package main

import (
	. "RemoteStudioLive/SharedUtils"
	"bufio"
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	"golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	StatisticsLog      = "StatisticsLog.txt"                                           // StatisticsLog - The file that logs the time measurements
	LogFile            = "log.txt"                                                     // LogFile - The file that is used for print and debug
	SongName           = "Eric Clapton - Nobody Knows You When You're Down & Out .mp3" // SongName - The song to send and play
	PacketRequestSong  = iota                                                          // PacketRequestSong - .
	PacketCloseChannel                                                                 // PacketCloseChannel - .
	PacketRecord                                                                       // PacketRecord - For recording a stream with microphone
)

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

func initChannels() (chan []int64, chan []byte, chan []byte, chan string, chan string) {
	statsChannel := make(chan []int64, BufferSize)
	streamChannel := make(chan []byte, bufio.MaxScanTokenSize)
	handleResponseChannel := make(chan []byte, bufio.MaxScanTokenSize)
	endSessionChannel := make(chan string, bufio.MaxScanTokenSize)
	logChannel := make(chan string, bufio.MaxScanTokenSize)
	return statsChannel, streamChannel, handleResponseChannel, endSessionChannel, logChannel
}

// mean calculates the mean value from a slice of int64.
func mean(values []int64) float64 {
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// Jitter calculates the jitter for a slice of int64.
func jitter(values []int64) float64 {
	quadDev := float64(0)
	mean := mean(values)
	for _, v := range values {
		quadDev += math.Pow(float64(v)-mean, 2)
	}
	quadDev = quadDev / float64(len(values))
	return math.Sqrt(quadDev)
}

// CalculateInterArrival compute the differences between consecutive elements in a byte slice using map and a lambda function
func CalculateInterArrival(input []int64) []int64 {
	var output []int64
	for i := 0; i < len(input)-1; i++ {
		output = append(output, input[i+1]-input[i])
	}
	return output
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

// PipeSongToMPG receives a byte slice and prints it to the screen s.t mpg123 will catch and stream it
func pipeSongToMPG(byteSlice []byte) {
	_, err := os.Stdout.Write(byteSlice)
	CheckError(err)
}

// PlotByteSlice plots the values of a byte slice.
func plotByteSlice(data []int64, figName string, title string, xLabel string, yLabel string) error {
	// Create a new plot
	p := plot.New()
	// Create a new scatter plotter
	var points plotter.XYs
	for i := 0; i < len(data); i += 100 {
		points = append(points, plotter.XY{X: float64(i), Y: float64(data[i])})
	}
	s, err := plotter.NewScatter(points)
	if err != nil {
		return err
	}
	// Set color of the circles to blue and make them bold
	s.GlyphStyle.Color = color.RGBA{R: 63, G: 127, B: 191, A: 255}
	s.GlyphStyle.Radius = vg.Points(3) // Adjust the radius to make circles bold
	// Add scatter plotter to the plot
	p.Add(s)

	// Add a line between circles
	line, err := plotter.NewLine(points)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{R: 200, G: 200, B: 200, A: 255} // Gray color
	line.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}   // Dash pattern
	line.Width = vg.Points(0.5)                             // Line width
	p.Add(line)

	// Set labels for axes
	p.Title.Text = title
	p.Title.TextStyle.Font.Weight = font.WeightBold
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = xLabel
	p.X.Label.TextStyle.Font.Weight = font.WeightBold
	p.Y.Label.Text = yLabel
	p.Y.Label.TextStyle.Font.Weight = font.WeightBold

	// Save the plot to a file with additional white space around the plot
	err = p.Save(14*vg.Inch, 6*vg.Inch, figName)
	return err
}

func logMessage(logChannel chan string, message string) {
	logChannel <- message
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
