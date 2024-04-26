// Client src code
package main

import (
	. "RemoteStudioLive/SharedUtils"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {

	connSpecs := InitConnSpecs(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	conn, err := net.Dial(connSpecs.Type, connSpecs.IP+":"+connSpecs.Port)
	CheckError(err)
	defer conn.Close()

	// Create channels parallel sending, receiving, streaming and collecting messages.
	statsChannel, streamChannel, handleResponseChannel, endSessionChannel, logChannel := initChannels()

	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	{
		go logRoutine(LogFile, logChannel, &waitGroup)
		go statsRoutine(StatisticsLog, statsChannel, logChannel, &waitGroup)
		go streamRoutine(streamChannel, logChannel, &waitGroup)
		go handleResponseRoutine(conn, streamChannel, statsChannel, endSessionChannel, logChannel, &waitGroup)
	}

	// Close resources and synchronize goroutines
	defer func() {
		close(endSessionChannel)
		close(handleResponseChannel)
		close(statsChannel)
		close(streamChannel)
		time.Sleep(4 * time.Minute)
		close(logChannel)

		// Wait for the goroutines to finish
		waitGroup.Wait()
	}()

	sendSong(conn, SongName, endSessionChannel, logChannel)
	log(logChannel, "Exit Code 0")
}

func sendSong(conn net.Conn, songFileName string, endSessionChannel chan string, logChannel chan string) {
	file, err := os.Open(songFileName) // open the song that the clients wants to send to the server
	CheckError(err)
	defer file.Close()
	defer log(logChannel, "sendSong Done")

	buffer := make([]byte, DataFrameSize)
	// Send the song to the server (as packets)
	for {
		tInit := time.Now().UnixMicro()
		//time.Sleep(100000)
		bytesRead, err := file.Read(buffer)

		if err != nil { // When reading EOF
			log(logChannel, "sendSong err:"+err.Error())
			packet := Packet{PacketType: PacketCloseChannel}
			packet.SendPacket(conn)
			break
		}
		tProcessing := time.Now().UnixMicro() - tInit
		songPacket := InitPacket(PacketRequestSong, tInit, tProcessing, bytesRead)
		songPacket.SetData(buffer)
		songPacket.SendPacket(conn)
	}
	// Wait until communication is done

	for {
		msg := <-endSessionChannel
		switch msg {
		case "endSession":
			log(logChannel, "endSessionChannel got 'endSession' ")
			//time.Sleep(3*time.Second)
			return

		default:
			log(logChannel, "endSessionChannel got an unexpected message")
		}
	}
}

func handleResponseRoutine(conn net.Conn, streamChannel chan []byte, statsChannel chan []int64, endSessionChannel chan string, logChannel chan string, waitGroup *sync.WaitGroup) {
	log(logChannel, "handleResponseRoutine Start")
	defer waitGroup.Done()
	defer log(logChannel, "handleResponseRoutine Done")
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
			statsChannel <- []int64{timeStampFinal, int64(receivePacket.ProcessingTime), roundTripTime}

		case PacketCloseChannel:
			endSessionChannel <- "endSession"
			log(logChannel, "handleResponseRoutine got 'endSession' message")
			return

		default:
			log(logChannel, "handleResponseRoutine got an unexpected message ")
		}
	}
}

func logRoutine(fileName string, logChannel chan string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	log(logChannel, "logRoutine Start")
	CheckError(deleteFile(fileName))
	logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	defer logFile.Close()

	for {
		logMessage, ok := <-logChannel
		_ = logMessage
		if !ok {
			// The channel has been closed
			break
		}
			fmt.Fprintln(logFile, logMessage)
	}
	fmt.Fprintf(logFile, "logRoutine Done")
}

func statsRoutine(fileName string, statsChannel chan []int64, logChannel chan string, waitGroup *sync.WaitGroup) {
	log(logChannel, "statsRoutine Start")
	defer waitGroup.Done()
	defer log(logChannel, "statsRoutine Done")
	var roundTripTimes []int64
	var processingTimes []int64
	var arrivalTimes []int64
	CheckError(deleteFile(fileName))
	statisticsFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	defer statisticsFile.Close()

	packetIndex := 0
	// Listen on the channel
	for {
		timeMeasures, ok := <-statsChannel
		if !ok {
			// The channel has been closed
			break
		}
		arrivalTime, roundTripTime, processingTime := timeMeasures[0], timeMeasures[1], timeMeasures[2]
		if packetIndex%10 == 0 {
			fmt.Fprintf(statisticsFile, "Packet %4d | Round Trip Time: %5d microseconds\n", packetIndex, roundTripTime)
		}

		roundTripTimes = append(roundTripTimes, roundTripTime)
		processingTimes = append(processingTimes, processingTime)
		arrivalTimes = append(arrivalTimes, arrivalTime)
		packetIndex++
	}

	interArrivals := CalculateInterArrival(arrivalTimes)
	meanInterArrivals := int(mean(interArrivals))
	meanSendingTime := int(mean(roundTripTimes))
	rttJitter := int(jitter(roundTripTimes))
	// Plot graphs and print to statistics file
	{
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
}

func streamRoutine(streamChannel chan []byte, logChannel chan string, waitGroup *sync.WaitGroup) {
	log(logChannel, "streamRoutine Start")
	defer waitGroup.Done()
	defer log(logChannel, "streamRoutine Done")
	for {
		chunk, ok := <-streamChannel
		if !ok {
			// The channel has been closed
			return
		}
		pipeSongToMPG(chunk)
		//_ = chunk
	}
}
