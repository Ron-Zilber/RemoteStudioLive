// Client src code
package main

import (
	. "RemoteStudioLive/SharedUtils"
	"fmt"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gordonklaus/portaudio"
	"layeh.com/gopus"
)

func main() {
	connSpecs := InitConnSpecs(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	frameSize, _ := strconv.Atoi(os.Args[5])

	conn, err := dial(connSpecs.Type, connSpecs.IP+":"+connSpecs.Port)
	CheckError(err)
	defer conn.Close()

	// Create channels parallel sending, receiving, streaming and collecting messages.
	statsChannel, streamChannel, handleResponseChannel, endSessionChannel, logChannel := initChannels()

	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	{
		go logRoutine(LogFile, logChannel, &waitGroup)
		logFiles := []string{StatisticsLog, InterArrivalLog}
		go statsRoutine(logFiles, statsChannel, logChannel, &waitGroup, frameSize)
		go streamRoutine(streamChannel, logChannel, &waitGroup, connSpecs.OpMode, frameSize)
		go handleResponseRoutine(conn, streamChannel, statsChannel, endSessionChannel, logChannel, &waitGroup)
	}

	// Close resources and synchronize goroutines
	defer func() {
		close(endSessionChannel)
		close(handleResponseChannel)
		close(statsChannel)
		close(streamChannel)
		if connSpecs.OpMode == "song" {
			time.Sleep(4 * time.Minute) // Wait until song playing is
		} else { // OpMode == "record"
			time.Sleep(3 * time.Second)
		}
		close(logChannel)

		// Wait for the goroutines to finish
		waitGroup.Wait()
	}()

	switch connSpecs.OpMode {
	case "song":
		sendSong(conn, SongName, endSessionChannel, logChannel)
	case "record":
		fmt.Println("Starting session with", getAudioLength(frameSize), "millisecond framesize")
		recordAndSend(conn, logChannel, endSessionChannel, 20, frameSize)
	}

	logMessage(logChannel, "Exit Code 0")
	fmt.Println("")
}

func sendSong(conn net.Conn, songFileName string, endSessionChannel, logChannel chan string) {
	file, err := os.OpenFile(songFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	CheckError(err)
	defer func() {
		file.Close()
		logMessage(logChannel, "sendSong Done")
	}()

	buffer := make([]byte, DataFrameSize)
	// Send the song to the server (as packets)
	packetsCounter := 0
	for {
		//time.Sleep(time.Millisecond)
		tInit := time.Now().UnixMilli()
		bytesRead, err := file.Read(buffer)

		if err != nil { // When reading EOF
			logMessage(logChannel, "sendSong err:"+err.Error())
			packet := Packet{PacketType: PacketCloseChannel}
			packet.SendPacket(conn)
			break
		}
		tProcessing := time.Now().UnixMilli() - tInit
		songPacket := InitPacket(PacketRequestSong, packetsCounter, tInit, tProcessing, bytesRead)
		songPacket.SetData(buffer)
		songPacket.SendPacket(conn)
		packetsCounter++
	}

	// Wait until communication is done
	for {
		msg := <-endSessionChannel
		switch msg {
		case "endSession":
			logMessage(logChannel, "endSessionChannel got 'endSession' ")
			return

		default:
			logMessage(logChannel, "endSessionChannel got an unexpected message")
		}
	}
}

func recordAndSend(conn net.Conn, logChannel, endSessionChannel chan string, durationSeconds, frameSize int) {
	logMessage(logChannel, "recordAndSend Start")
	defer logMessage(logChannel, "recordAndSend Done")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	portaudio.Initialize()
	defer portaudio.Terminate()
	audioBufferSize := frameSize * Channels
	in := make([]int16, audioBufferSize)
	stream, err := portaudio.OpenDefaultStream(Channels, 0, SampleRate, 1, in)
	CheckError(err)
	defer stream.Close()

	encoder, err := gopus.NewEncoder(SampleRate, Channels, gopus.Audio)
	CheckError(err)
	tInit := time.Now().UnixMicro()
	CheckError(stream.Start())

	packetsCounter := 0
	fmt.Println("Record start")
	for {
		tRecordFrame := time.Now().UnixMicro()
		//time.Sleep(10*time.Millisecond)
		CheckError(stream.Read())                                   //* Read filling the buffer by recording samples until the buffer is full
		data, err := encoder.Encode(in, frameSize, audioBufferSize) //* Encode PCM to Opus
		if err != nil {
			logMessage(logChannel, "recordAndSend error: "+err.Error())
			break
		}
		tProcessing := time.Now().UnixMicro() - tRecordFrame
		recordPacket := InitPacket(PacketRecord, packetsCounter, tRecordFrame, tProcessing, len(data))
		packetsCounter++
		recordPacket.SetData(data)
		recordPacket.SendPacket(conn)

		select {
		case <-sig:
			CheckError(stream.Stop())
			packet := Packet{PacketType: PacketCloseChannel}
			packet.SendPacket(conn)
			return

		default:
			if time.Now().UnixMicro()-tInit > int64(durationSeconds)*MicroToSecond {
				fmt.Println("Record end")
				logMessage(logChannel, "recordAndPlay Timeout")
				packet := Packet{PacketType: PacketCloseChannel}
				packet.SendPacket(conn)
				CheckError(stream.Stop())
				return
			}
		}
	}

	// Wait until communication is done
	<-endSessionChannel
	logMessage(logChannel, "endSessionChannel got 'endSession' ")
}

func handleResponseRoutine(conn net.Conn, streamChannel chan []byte, statsChannel chan []int64, endSessionChannel, logChannel chan string, waitGroup *sync.WaitGroup) {
	logMessage(logChannel, "handleResponseRoutine Start")
	defer waitGroup.Done()
	defer logMessage(logChannel, "handleResponseRoutine Done")

	for {
		var receivePacket Packet
		receivePacket.ReadPacket(conn)

		switch receivePacket.PacketType {

		case PacketRequestSong, PacketRecord:
			timeStampFinal := time.Now().UnixMicro()
			endToEnd := timeStampFinal - int64(receivePacket.InitTime)
			statsChannel <- []int64{
				int64(receivePacket.SerialNumber),
				timeStampFinal,
				int64(receivePacket.ProcessingTime),
				endToEnd,
			}
			streamChannel <- receivePacket.Data[:receivePacket.DataSize]

		case PacketCloseChannel:
			endSessionChannel <- "endSession"
			logMessage(logChannel, "handleResponseRoutine got 'endSession' message")
			return

		default:
			logMessage(logChannel, "handleResponseRoutine got an unexpected message ")
		}
	}
}

func logRoutine(fileName string, logChannel chan string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	logMessage(logChannel, "logRoutine Start")
	var logBuffer strings.Builder

	for {
		logMessage, ok := <-logChannel
		if !ok {
			// The channel has been closed
			break
		}
		logBuffer.WriteString(logMessage + "\n")
	}
	// Export results to file
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	CheckError(err)
	defer logFile.Close()
	logBuffer.WriteString("logRoutine Done\n")
	fmt.Fprint(logFile, logBuffer.String())
}

func statsRoutine(fileNames []string, statsChannel chan []int64, logChannel chan string, waitGroup *sync.WaitGroup, frameSize int) {
	logMessage(logChannel, "statsRoutine Start")
	defer waitGroup.Done()
	defer logMessage(logChannel, "statsRoutine Done")

	statisticsFileName := strings.TrimSuffix(fileNames[0], ".txt") + " " + strconv.Itoa(frameSize) + ".txt"
	interArrivalFileName := strings.TrimSuffix(fileNames[1], ".txt") + " " + strconv.Itoa(frameSize) + ".txt"

	var (
		serialNumbers, endToEnds, roundTripTimes, arrivalTimes []int64
		statisticsBuffer                                       strings.Builder
	)
	// Listen on the channel
	for {
		timeMeasures, ok := <-statsChannel
		if !ok {
			// The channel has been closed
			break
		}
		serialNumber, arrivalTime := timeMeasures[0], timeMeasures[1]
		processingTime, endToEnd := timeMeasures[2], timeMeasures[3]
		roundTripTime := endToEnd - processingTime
		infoString := fmt.Sprintf(
			"Packet %4d | End To End: %5d microseconds | Round Trip Time: %4d microseconds\n",
			serialNumber, endToEnd, roundTripTime)

		statisticsBuffer.WriteString(infoString)
		serialNumbers = append(serialNumbers, serialNumber)
		endToEnds = append(endToEnds, endToEnd)
		roundTripTimes = append(roundTripTimes, roundTripTime)
		arrivalTimes = append(arrivalTimes, arrivalTime)
	}
	// Export results to file
	statisticsFile, err := os.OpenFile(statisticsFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	CheckError(err)
	defer statisticsFile.Close()

	fmt.Fprint(statisticsFile, statisticsBuffer.String())
	interArrivals := CalculateInterArrival(arrivalTimes)

	interArrivalFile, err := os.OpenFile(interArrivalFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	CheckError(err)
	defer interArrivalFile.Close()
	fmt.Fprintln(interArrivalFile, int64sToString(interArrivals))
	meanInterArrivals := mean(interArrivals)
	meanEndToEnd := mean(endToEnds)
	meanRoundTripTime := mean(roundTripTimes)
	rttJitter := jitter(roundTripTimes) // TODO: Should the jitter be calculated on end to end or rtt?

	unordered := countUnordered(serialNumbers)
	lostPackets := countLostPackets(serialNumbers)
	sentPackets := slices.Max(serialNumbers) + 1

	unorderedPercentage := getPercentage(int(unordered), sentPackets)
	lostPacketsPercentage := getPercentage(lostPackets, sentPackets)

	//fmt.Println("Unordered packets:", unordered, " Out of", sentPackets, " Packets", unorderedPercentage, "%")
	//fmt.Println("Lost packets:", lostPackets, " Out of", sentPackets, " Packets", lostPacketsPercentage, "%")

	metrics := NetworkMetrics{
		frameSize:         getAudioLength(frameSize),
		endToEnd:          toMilli(meanEndToEnd),
		roundTripTime:     toMilli(meanRoundTripTime),
		interArrival:      toMilli(meanInterArrivals),
		jitter:            toMilli(rttJitter),
		unorderedArrivals: unorderedPercentage,
		lostPackets:       lostPacketsPercentage,
	}

	CheckError(updateStats(SummarizedStatsFile, &metrics))

	// Plot graphs and print to statistics file
	/*
		{
			fmt.Fprint(statisticsFile, "\n") // Add an empty line

			fmt.Fprintln(statisticsFile,
				"Average Round Trip Time:        ", meanEndToEnd, "milliseconds")

			fmt.Fprintln(statisticsFile,
				"Round Trip Time Jitter:         ", rttJitter, "milliseconds")

			fmt.Fprintln(statisticsFile,
				"Average Inter-Arrival Time:     ", meanInterArrivals, "milliseconds")

			CheckError(plotByteSlice(endToEnds,
				"./Plots/Packets RTT Plot.png",
				"Packets RTT [milliseconds]",
				"Packet Index",
				"Packet RTT [milliseconds]"))

			CheckError(plotByteSlice(interArrivals,
				"./Plots/Inter-Arrival Times.png",
				"Inter-Arrival Times [milliseconds]",
				"Packet Index",
				"Inter-Arrival Time [milliseconds]"))
		}
	*/
}

func streamRoutine(streamChannel chan []byte, logChannel chan string, waitGroup *sync.WaitGroup, workMode string, frameSize int) {
	logMessage(logChannel, "streamRoutine Start")
	defer func() {
		waitGroup.Done()
		logMessage(logChannel, "streamRoutine Done")
	}()
	switch workMode {
	case "song":
		for {
			chunk, ok := <-streamChannel
			if !ok {
				// The channel has been closed
				return
			}
			pipeSongToMPG(chunk)
		}

	case "record":
		decoder, err := gopus.NewDecoder(SampleRate, Channels)
		CheckError(err)
		audioBufferSize := frameSize * Channels
		CheckError(speaker.Init(beep.SampleRate(SampleRate), audioBufferSize))
		var buffer [][2]float64

		streamer := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
			if len(buffer) == 0 {
				chunk, ok := <-streamChannel
				if !ok {
					// The channel has been closed
					return 0, false
				}
				pcm, err := decoder.Decode(chunk, frameSize, false)
				if err != nil {
					logMessage(logChannel, "Error in streamRoutine: "+err.Error())
					return 0, false
				}

				for i := 0; i < len(pcm); i += 2 {
					buffer = append(buffer, [2]float64{
						float64(pcm[i]) / 32768.0,
						float64(pcm[i+1]) / 32768.0,
					})
				}
			}

			for i := range samples {
				if len(buffer) == 0 {
					return i, true
				}
				samples[i] = buffer[0]
				buffer = buffer[1:]
			}
			return len(samples), true
		})

		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))

		<-done
	}
}
