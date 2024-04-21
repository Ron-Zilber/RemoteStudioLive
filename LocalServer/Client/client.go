// Client src code
package main

import (
	. "RemoteStudioLive/SharedUtils"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ServerIP          = "172.23.175.237"                                              // ServerIP   - Replace with the actual IP address of your server
	ServerPort        = "8080"                                                        // ServerPort - The port number of the server
	ConnType          = "tcp"                                                         // ConnType   - The type of the connection
	OpMode            = "default"                                                     // OpMode - The operation mode of the client
	StatisticsLog     = "StatisticsLog.txt"                                           // StatisticsLog - The file that logs the time measurements
	SongName          = "Eric Clapton - Nobody Knows You When You're Down & Out .mp3" // SongName - The song to send and play
	PacketRequestSong = iota                                                          // PacketRequestSong - .
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
	CheckError(err)
	defer conn.Close()
	// Create a channel to receive statistics
	statsChannel := make(chan []int64, BufferSize)
	streamChannel := make(chan []byte, bufio.MaxScanTokenSize)
	handleResponseChannel := make(chan []byte, bufio.MaxScanTokenSize)

	// Create a wait group to synchronize goroutines
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	go statsRoutine(StatisticsLog, statsChannel, &waitGroup)
	go streamRoutine(streamChannel, &waitGroup)
	go handleResponseRoutine(conn, streamChannel, statsChannel, &waitGroup)

	// Start to transmit media in according to the opMode
	for {
		if !sendMessage(conn, opMode) {
			break
		}
	}
	// Close the channels to signal the goroutines to exit
	close(statsChannel)
	close(streamChannel)
	close(handleResponseChannel)
	// Wait for the goroutines to finish
	waitGroup.Wait()
}

func sendMessage(conn net.Conn, opMode string) bool {
	reader := bufio.NewReader(os.Stdin)
	switch strings.TrimSpace(opMode) {
	case "song":
		sendSong(conn, SongName)
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
			sendFile(conn, "lorem ipsum.txt")
			return true

		case "time":
			timeStamp := time.Now().UnixMicro()
			fmt.Fprintf(conn, "Time: "+strconv.Itoa(int(timeStamp))+"\n")

		case "ping":
			pingServer(conn)
		default:
			fmt.Fprintf(conn, text+"\n")
		}
	}
	return true
}
func sendSong(conn net.Conn, songFileName string) {
	file, err := os.Open(songFileName) // open the song that the clients wants to send to the server
	CheckError(err)
	// close file on exit and check for its returned error
	defer file.Close()
	// make a buffer to keep chunks that are read

	buffer := make([]byte, DataFrameSize)
	for {
		tInit := time.Now().UnixMicro()
		// TODO: Add sleep()
		bytesRead, err := file.Read(buffer)
		
		if err != nil {
			break
		}
		tProcessing := time.Now().UnixMicro() - tInit
		songPacket := ConstructPacket(PacketRequestSong, tInit, tProcessing, bytesRead)
		songPacket.SetData(buffer)
		songPacket.SendPacket(conn)
	}
	return
}

func handleResponseRoutine(conn net.Conn, streamChannel chan []byte, statsChannel chan []int64, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		var receivePacket Packet
		receivePacket.ReadPacket(conn)
		switch receivePacket.PacketType {
		case PacketRequestSong:
			chunk := make([]byte, receivePacket.DataSize)
			copy(chunk, receivePacket.Data[:])

			// Send the packet to the routine that pipelines the packets to mpg123
			streamChannel <- chunk

			timeStampFinal := time.Now().UnixMicro()
			roundTripTime := timeStampFinal - int64(receivePacket.InitTime) - int64(receivePacket.ProcessingTime)
			statsChannel <- []int64{int64(receivePacket.InitTime), int64(receivePacket.ProcessingTime), roundTripTime}
		}
	}
}
func statsRoutine(fileName string, statsChannel chan []int64, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var roundTripTimes []int64
	var processingTimes []int64
	var initTimes []int64
	deleteFile(fileName)
	statisticsFile, err := os.OpenFile(StatisticsLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	defer statisticsFile.Close()

	packetIndex := 0
	for timeMeasures := range statsChannel {
		initTime, roundTripTime, processingTime := timeMeasures[0], timeMeasures[1], timeMeasures[2]
		if packetIndex%10 == 0 {
			fmt.Fprintf(statisticsFile, "Packet %4d | Round Trip Time: %5d microseconds\n", packetIndex, roundTripTime)
		}

		roundTripTimes = append(roundTripTimes, roundTripTime)
		processingTimes = append(processingTimes, processingTime)
		initTimes = append(initTimes, initTime)
		packetIndex++
	}

	interArrivals := CalculateInterArrival(initTimes)
	meanInterArrivals := int(mean(interArrivals))
	meanSendingTime := int(mean(roundTripTimes))
	rttJitter := int(jitter(roundTripTimes))

	fmt.Fprint(statisticsFile, "\n") // Add an empty line
	fmt.Fprintln(statisticsFile, "Average Round Trip Time:        ", meanSendingTime, "microseconds")
	fmt.Fprintln(statisticsFile, "Round Trip Time Jitter:         ", rttJitter, "microseconds")
	fmt.Fprintln(statisticsFile, "Average Inter-Arrival Time:     ", meanInterArrivals, "microseconds")
	CheckError(plotByteSlice(roundTripTimes,
		"Packets RTT Plot.png",
		"Packets RTT [microseconds]",
		"Packet Index",
		"Packet RTT [microseconds]"))

	CheckError(plotByteSlice(interArrivals,
		"Inter-Arrival Times.png",
		"Inter-Arrival Times [microseconds]",
		"Packet Index",
		"Inter-Arrival Time [microseconds]"))
}

func streamRoutine(streamChannel chan []byte, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		chunk, ok := <-streamChannel
		if !ok {
			// The channel has been closed
			return
		}
		pipeSongToMPG(chunk)
	}
}

func handleRequest(conn net.Conn, streamChannel chan []byte) {
	reader := bufio.NewReader(conn)
	var receivePacket Packet
	receivePacket.ReadPacket(conn)

	switch receivePacket.PacketType {
	case PacketRequestSong:
		chunk := make([]byte, receivePacket.DataSize)
		copy(chunk, receivePacket.Data[:])
		streamChannel <- chunk

	default:
		message, err := reader.ReadString('\n')
		CheckError(err)
		fmt.Print("Message from the server: " + message)

		if len(message) >= len("TIME") && message[:len("TIME")] == "TIME" {
			parts := strings.Fields(message)
			oldTimeStamp, err := strconv.Atoi(parts[1])
			CheckError(err)
			newTimeStamp := time.Now().UnixMicro()
			elapsedTime := newTimeStamp - int64(oldTimeStamp)
			fmt.Println("Elapsed time:", elapsedTime, "Microseconds")
		}
	}
}

func sendFile(conn net.Conn, fileName string) {
	file, err := os.Open(fileName)
	CheckError(err)
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
			handleRequest(conn, nil)
			s, err = reader.ReadString('\n')
		}
		// Send the last chunk
		fmt.Fprintf(conn, s+"\n")
		break
	}
}

func pingServer(conn net.Conn) {
	i, count, frequency := 1, 100, 50*time.Millisecond
	for i < count {
		fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n")
		handleRequest(conn, nil)
		time.Sleep(frequency)
		i++
	}
	fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n") // Send the last ping
}
