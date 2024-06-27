package main

import (
	. "RemoteStudioLive/SharedUtils"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gordonklaus/portaudio"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	RttLog             = "./Stats/StatisticsLog.txt"                                   // StatisticsLog - The file that logs the time measurements
	InterArrivalLog    = "./Stats/interArrivalLog.txt"                                 // InterArrivalLog - The file that logs the inter-arrivals
	LogFile            = "log.txt"                                                     // LogFile - The file that is used for print and debug
	SongName           = "Eric Clapton - Nobody Knows You When You're Down & Out .mp3" // SongName - The song to send and play
	PacketRequestSong  = iota                                                          // PacketRequestSong - .
	PacketCloseChannel                                                                 // PacketCloseChannel - .
	PacketRecord                                                                       // PacketRecord - For recording a stream with microphone
	SampleRate         = 48000                                                         // SampleRate is the number of bits used to represent a full second of audio sampling
	Channels           = 2                                                             // Channels - 1 for mono; 2 for stereo
	FrameSize          = 960                                                           // FrameSize of 960 gives 20 ms (for 48kHz sampling) which is the Opus recommendation
	AudioBufferSize    = FrameSize * Channels                                          // AudioBufferSize let the buffer hold multiple frames
)

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

func dial(netType string, address string) (net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)
	switch netType {
	case "tcp":
		conn, err = net.Dial(netType, address)
	case "udp":
		udpConn, err := net.ResolveUDPAddr("udp", address)
		CheckError(err)

		conn, err = net.DialUDP("udp", nil, udpConn)
		return conn, err
	}
	return conn, err
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

func recordRawToBytes(milliSeconds int64) ([]byte, error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

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
	tStart := time.Now().UnixMilli()

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
			if time.Now().UnixMilli()-tStart > milliSeconds {
				break loop
			}
		}
	}
	if err := stream.Stop(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// StreamFileToMPG writes []bytes to temporary
// file, streaming the audio and deleting the file
func streamFileToMPG(chunk []byte) {
	tempName := time.Now().String() + ".mp3"
	encodeBytesToMp3(chunk, tempName)
	cmd := exec.Command("mpg123", " - ", tempName)
	CheckError(cmd.Start())
	cmd.Wait()
	CheckError(os.Remove(tempName))
	//fmt.Println("streamFileToMPG done")
}

func int64sToString(list []int64) string {
	var s strings.Builder
	for _, num := range list {
		s.WriteString(strconv.Itoa(int(num)) + "\n")
	}
	return s.String()
}
