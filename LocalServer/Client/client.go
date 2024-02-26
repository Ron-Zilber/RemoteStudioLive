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
	"time"
)

const (
	ServerIP       = "172.23.175.237"     // ServerIP   - Replace with the actual IP address of your server
	ServerPort     = "8080"               // ServerPort - The port number of the server
	ConnType       = "tcp"                // ConnType   - The type of the connection
	OpMode         = "default"            // OpMode - The operation mode of the client
	SongFromClient = "SongFromClient.mp3" // SongFromClient - The song that echoed back from the server
	StatisticsLog  = "StatisticsLog.txt"  // StatisticsLog - The file that logs the time measurements
)

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
	deleteFile(SongFromClient)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		if !sendMessage(conn, opMode) {
			break
		}
		handleRequest(conn, opMode)
	}

	if opMode == "song" {
		pipeSongToMPG()
	}
}

func sendMessage(conn net.Conn, opMode string) (status bool) {
	reader := bufio.NewReader(os.Stdin)
	switch strings.TrimSpace(opMode) {
	case "song":
		sendSong(conn, "Eric Clapton - Nobody Knows You When You're Down & Out .mp3", opMode)
		return false

	case "measure":
		deleteFile(StatisticsLog)
		statisticsFile, err := os.OpenFile(StatisticsLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		checkError(err)
		defer statisticsFile.Close()
		count := 100
		timeMeasures := make([]int64, count)
		for i := 0; i < count; i++ {
			deleteFile(SongFromClient)
			timeMeasures[i] = sendSong(conn, "Eric Clapton - Nobody Knows You When You're Down & Out .mp3", opMode)
			fmt.Fprint(statisticsFile, strconv.Itoa(i+1)+". Elapsed time:"+strconv.Itoa(int(timeMeasures[i]))+" Microseconds\n")
		}
		meanSendingTime := Mean(timeMeasures)
		jitter := Jitter(timeMeasures)
		fmt.Fprintln(statisticsFile, "") // Add an empty line
		fmt.Fprintln(statisticsFile, "Average elapsed time: ", meanSendingTime, "Microseconds")
		fmt.Fprintln(statisticsFile, "Jitter of the elapsed time: ", jitter, "Microseconds")
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

func handleRequest(conn net.Conn, opMode string) {
	reader := bufio.NewReader(conn)
	switch strings.TrimSpace(opMode) {

	case "measure":
		songFile, err := os.OpenFile(SongFromClient, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		checkError(err)
		defer songFile.Close()
		buffer := make([]byte, bufio.MaxScanTokenSize)
		// Read chunk
		bytesRead, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		// Save chunk to the file
		_, err = songFile.Write(buffer[:bytesRead])
		checkError(err)
	case "song":
		songFile, err := os.OpenFile(SongFromClient, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		checkError(err)
		defer songFile.Close()
		buffer := make([]byte, bufio.MaxScanTokenSize)
		// Read chunk
		bytesRead, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		// Save chunk to the file
		_, err = songFile.Write(buffer[:bytesRead])
		checkError(err)

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
func sendSong(conn net.Conn, fileName string, opMode string) (timeMeasure int64) {
	file, err := os.Open(fileName)
	timeStampInitial := time.Now().UnixMicro()
	checkError(err)
	// close file on exit and check for its returned error
	defer file.Close()
	// make a buffer to keep chunks that are read
	buffer := make([]byte, bufio.MaxScanTokenSize)
	for {
		// Read chunk
		bytesReads, err := file.Read(buffer)
		for err == nil { // When EOF will be read, err != nil
			// Send chunk

			_, err = conn.Write(buffer[:bytesReads])
			handleRequest(conn, opMode)
			bytesReads, err = file.Read(buffer)
		}
		// Send the last chunk
		_, err = conn.Write(buffer[:bytesReads])
		break
	}
	timeStampFinal := time.Now().UnixMicro()
	elapsedTime := timeStampFinal - timeStampInitial
	return elapsedTime
}

func pingServer(conn net.Conn, opMode string) {
	i := 1
	count := 100
	frequency := 100 * time.Millisecond
	for i < count {
		fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n")
		handleRequest(conn, opMode)
		time.Sleep(frequency)
		i++
	}
	fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n") // Send the last ping
}
func pipeSongToMPG() {
	mp3Data, err := os.ReadFile(SongFromClient)
	checkError(err)
	// Write the MP3 data to stdout
	_, err = os.Stdout.Write(mp3Data)
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
	var quadDev float64
	mean := Mean(values)
	for _, v := range values {
		quadDev += math.Pow(float64(v)-mean, 2)
	}
	quadDev = quadDev / float64(len(values))
	return math.Sqrt(quadDev)
}
