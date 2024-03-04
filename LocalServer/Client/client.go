// Client src code
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ServerIP      = "172.23.175.237"                                              // ServerIP   - Replace with the actual IP address of your server
	ServerPort    = "8080"                                                        // ServerPort - The port number of the server
	ConnType      = "tcp"                                                         // ConnType   - The type of the connection
	OpMode        = "default"                                                     // OpMode - The operation mode of the client
	StatisticsLog = "StatisticsLog.txt"                                           // StatisticsLog - The file that logs the time measurements
	SongName      = "Eric Clapton - Nobody Knows You When You're Down & Out .mp3" // SongName - The song to send and play
	BufferSize    = bufio.MaxScanTokenSize / 32                                   // BufferSize - The size of the packets when transmitting a song
)

// Global Variables
var songByteSlice []byte

func main() {
	// Enable server IP and port configuring from shell
	serverPort, serverIP := ServerPort, ServerIP
	opMode := OpMode
	if len(os.Args) > 3 {
		serverIP = os.Args[1]
		serverPort = os.Args[2]
		opMode = os.Args[3]
	}
	// Connect to the server
	conn, err := net.Dial(ConnType, serverIP+":"+serverPort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// Create a channel to receive statistics
	statsChannel := make(chan string, 1024)
	// Create a wait group to synchronize goroutines
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	// Congruently run the statistics server
	go statsServer(StatisticsLog, statsChannel, &waitGroup)

	// Start to transmit media in according to the opMode
	for {
		if !sendMessage(conn, statsChannel, opMode) {
			break
		}
		handleRequest(conn, opMode)
	}

	// Close the channel to signal the goroutine to exit
	close(statsChannel)
	// Wait for the goroutine to finish
	waitGroup.Wait()
}

func sendMessage(conn net.Conn, statisticsChannel chan string, opMode string) bool {
	reader := bufio.NewReader(os.Stdin)
	switch strings.TrimSpace(opMode) {
	case "song":
		sendSong(conn, statisticsChannel, SongName, opMode)
		return false

	default:
		fmt.Print("Text to send: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			return true // Continue connection
		}
		switch strings.TrimSpace(text) {
		case "exit":
			fmt.Println(conn.LocalAddr().String(), "Disconnecting")
			fmt.Fprintf(conn, text+"\n")
			return false // Close connection

		case "lorem":
			sendFile(conn, "lorem ipsum.txt", opMode)
			return true

		case "time":
			timeStamp := time.Now().UnixMicro()
			fmt.Fprintf(conn, "Time: "+strconv.Itoa(int(timeStamp))+"\n")

		case "ping":
			pingServer(conn, opMode)
		default:
			fmt.Fprintf(conn, text+"\n")
		}
	}
	return true
}
func sendSong(conn net.Conn, statisticsChannel chan<- string, songFileName string, opMode string) {
	file, err := os.Open(songFileName) // open the song that the clients wants to send to the server
	checkError(err)
	// close file on exit and check for its returned error
	defer file.Close()
	// make a buffer to keep chunks that are read
	buffer := make([]byte, BufferSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			break
		}
		timeStampInitial := time.Now().UnixMilli()
		_, err = conn.Write(buffer[:bytesRead])
		handleRequest(conn, opMode)
		timeStampFinal := time.Now().UnixMilli()
		elapsedTime := timeStampFinal - timeStampInitial
		statisticsChannel <- strconv.Itoa(int(elapsedTime))
	}
	return
}

func handleRequest(conn net.Conn, opMode string) {
	reader := bufio.NewReader(conn)
	switch strings.TrimSpace(opMode) {

	case "song":
		buffer := make([]byte, bufio.MaxScanTokenSize)
		// Read chunk
		bytesRead, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		// Save chunk to the byteSlice
		chunk := buffer[:bytesRead]
		//songByteSlice = append(songByteSlice, chunk...)
		pipeSongToMPG(chunk)

	default:
		message, err := reader.ReadString('\n')
		checkError(err)
		fmt.Print("Message from the server: " + message)

		if len(message) >= len("TIME") && message[:len("TIME")] == "TIME" {
			parts := strings.Fields(message)
			oldTimeStamp, err := strconv.Atoi(parts[1])
			checkError(err)
			newTimeStamp := time.Now().UnixMicro()
			elapsedTime := newTimeStamp - int64(oldTimeStamp)
			fmt.Println("Elapsed time:", elapsedTime, "Microseconds")
		}
	}
}

func sendFile(conn net.Conn, fileName string, opMode string) {
	file, err := os.Open(fileName)
	checkError(err)
	// close file on exit and check for its returned error
	defer file.Close()
	// make a buffer to keep chunks that are read
	reader := bufio.NewReader(file)
	for {
		// read a chunk
		s, err := reader.ReadString('\n')
		for err == nil {
			// send a chunk
			fmt.Fprintf(conn, s+"\n")
			handleRequest(conn, opMode)
			s, err = reader.ReadString('\n')
		}
		// Send the last chunk
		fmt.Fprintf(conn, s+"\n")
		break
	}
}

func pingServer(conn net.Conn, opMode string) {
	i, count, frequency := 1, 100, 50*time.Millisecond
	for i < count {
		fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n")
		handleRequest(conn, opMode)
		time.Sleep(frequency)
		i++
	}
	fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n") // Send the last ping
}
func pipeSongToMPG(byteSlice []byte) {
	_, err := os.Stdout.Write(byteSlice)
	checkError(err)
}

// checkError General error handling
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
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

func statsServer(fileName string, statsChannel chan string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var timeMeasures []int64
	deleteFile(fileName)
	statisticsFile, err := os.OpenFile(StatisticsLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	checkError(err)
	defer statisticsFile.Close()
	for measure := range statsChannel {
		fmt.Fprintln(statisticsFile, "Delay: ", measure, " milliseconds")
		measureToInt, _ := strconv.ParseInt(measure, 10, 64)
		timeMeasures = append(timeMeasures, measureToInt)
	}
	//timeMeasures
	meanSendingTime := Mean(timeMeasures)
	jitter := Jitter(timeMeasures)
	fmt.Fprint(statisticsFile, "\n") // Add an empty line
	fmt.Fprintln(statisticsFile, "Average elapsed time: ", meanSendingTime, "milliseconds")
	fmt.Fprintln(statisticsFile, "Jitter of the elapsed time: ", jitter, "milliseconds")
}

// Mean calculates the mean value from a slice of int64.
func Mean(values []int64) float64 {
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// Jitter calculates the jitter for a slice of int64.
func Jitter(values []int64) float64 {
	quadDev := float64(0)
	mean := Mean(values)
	for _, v := range values {
		quadDev += math.Pow(float64(v)-mean, 2)
	}
	quadDev = quadDev / float64(len(values))
	return math.Sqrt(quadDev)
}
