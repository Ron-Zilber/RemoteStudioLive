package main

import (
	. "RemoteStudioLive/SharedUtils"
	"bufio"
	"image/color"
	"math"
	"os"

	"golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	ServerPort         = "8080"                                                        // ServerPort - The port number of the server
	ConnType           = "tcp"                                                         // ConnType   - The type of the connection
	OpMode             = "default"                                                     // OpMode - The operation mode of the client
	StatisticsLog      = "StatisticsLog.txt"                                           // StatisticsLog - The file that logs the time measurements
	SongName           = "Eric Clapton - Nobody Knows You When You're Down & Out .mp3" // SongName - The song to send and play
	PacketRequestSong  = iota                                                          // PacketRequestSong - .
	PacketCloseChannel                                                                 // PacketCloseChannel - .
)

func initChannels() (chan []int64, chan []byte, chan []byte, chan string) {
	statsChannel := make(chan []int64, BufferSize)
	streamChannel := make(chan []byte, bufio.MaxScanTokenSize)
	handleResponseChannel := make(chan []byte, bufio.MaxScanTokenSize)
	endSessionChannel := make(chan string, bufio.MaxScanTokenSize)
	return statsChannel, streamChannel, handleResponseChannel, endSessionChannel
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