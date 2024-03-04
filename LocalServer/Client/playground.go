/*
* A playground use to test and develop helper functions and operations.
* Not in used for 
*/

package main

import (
	"bufio"
	"image/color"
	"math/rand"
	"time"

	"golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	BufferSize1 = bufio.MaxScanTokenSize / 128 // BufferSize1 - The size of the packets when transmitting a song
)

// Global Variables
var songByteSlice1 []byte

func main1() {
	randomSlice := generateRandomInt64Slice(16000)
	plotByteSlice1(randomSlice)
}

// PlotByteSlice1 plots the values of a byte slice.

func plotByteSlice1(data []int64) error {
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
	line.Width = vg.Points(0.5)                              // Line width
	p.Add(line)

	// Set labels for axes
	p.Title.Text = "Packets Delay [milliseconds]"
	p.Title.TextStyle.Font.Weight = font.WeightBold
	p.Title.TextStyle.Font.Size = 14
	p.X.Label.Text = "Packet Index"
	p.X.Label.TextStyle.Font.Weight = font.WeightBold
	p.Y.Label.Text = "Packet Delay [milliseconds]"
	p.Y.Label.TextStyle.Font.Weight = font.WeightBold

	// Save the plot to a file with additional white space around the plot
	err = p.Save(14*vg.Inch, 6*vg.Inch, "byte_slice_plot.png")
	return err
}


func generateRandomInt64Slice(size int) []int64 {
	rand.Seed(time.Now().UnixNano())

	// Create a slice to hold the generated random numbers
	randomSlice := make([]int64, size)

	// Generate random numbers and fill the slice
	for i := 0; i < size; i++ {
		randomSlice[i] = int64(rand.Intn(131)) // Generate random number in the range [0, 130]
	}

	return randomSlice
}
