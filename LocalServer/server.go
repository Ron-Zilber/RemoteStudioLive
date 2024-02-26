// Server src code
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	ConnHost = ""        // ConnHost - Empty string means listen on all available interfaces
	ConnPort = "8080"    // ConnPort - The port of the connection
	ConnType = "tcp"     // ConnType - The type of the connection
	OpMode   = "default" // OpMode - The operation mode
)

func main() {
	//plotFoo()
	//os.Exit(0)
	opMode := OpMode
	// Enable port configuring from shell
	connPort := ConnPort
	if len(os.Args) > 2 {
		connPort = os.Args[1]
		opMode = os.Args[2]
	}
	// Listen for incoming connection.
	ln, err := net.Listen(ConnType, ":"+connPort)
	CheckError(err)
	// Close the listener when the application closes.
	defer ln.Close()
	fmt.Println("Listening on port:", connPort)

	// Listen for an incoming connection.
	for {
		conn, err := ln.Accept()
		fmt.Println("Connected to:", conn.RemoteAddr().String())
		CheckError(err)
		// Handle incoming messages
		go handleConnection(conn, opMode)
	}
}

func handleConnection(conn net.Conn, opMode string) {
	// Handle incoming messages
	defer conn.Close()
	reader := bufio.NewReader(conn)
	switch strings.TrimSpace(opMode) {
	case "song":
		buffer := make([]byte, bufio.MaxScanTokenSize)
		for {
			// Read chunk
			bytesRead, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				} else {
					fmt.Println(conn.RemoteAddr(), "Disconnected")
					break
				}
				break
			}
			// Send chunk back to the client
			_, err = conn.Write(buffer[:bytesRead])
			CheckError(err)
		}
	default:
		for {
			message, err := reader.ReadString('\n')
			if message == "exit\n" || err == io.EOF {
				fmt.Println(conn.RemoteAddr(), "Disconnected")
				break
			}
			CheckError(err)
			fmt.Print("Message Received from " + conn.RemoteAddr().String() + " " + string(message))
			newMessage := strings.ToUpper(message)
			conn.Write([]byte(newMessage))
		}
	}
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func plotFoo() {
	// Generate sample data
	x := make([]float64, 100)
	y := make([]float64, 100)
	for i := range x {
		x[i] = float64(i)
		y[i] = float64(i) + rand.Float64()*10.0 // Random data for demonstration
	}

	// Plot the data
	if err := PlotXY(x, y, "output.png", "Title", "X-axis Label", "Y-axis Label"); err != nil {
		log.Fatal(err)
	}
}

// PlotXY generates a plot of y as a function of x and saves it to a file.
func PlotXY(x, y []float64, filename, title, xlabel, ylabel string) error {
	p := plot.New()

	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel

	points := make(plotter.XYs, len(x))
	for i := range points {
		points[i].X = x[i]
		points[i].Y = y[i]
	}

	err := plotutil.AddLinePoints(p, "Data", points)
	if err != nil {
		return err
	}

	// Save the plot to a file
	if err := p.Save(10*vg.Inch, 6*vg.Inch, filename); err != nil {
		return err
	}
	return nil
}
