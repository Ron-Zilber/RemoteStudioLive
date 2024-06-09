// Client src code
package main

import (
	. "RemoteStudioLive/SharedUtils"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gordonklaus/portaudio"
	"layeh.com/gopus"
)

const ()

func main() {
	workMode := "record" // TODO: change this approach of choosing between live record or streaming file
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
		go streamRoutine(streamChannel, logChannel, &waitGroup, workMode)
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

	switch workMode { // TODO: Change this mechanism to command line argument
	case "song":
		sendSong(conn, SongName, endSessionChannel, logChannel)
	case "record":
		//sendRecord(conn, endSessionChannel, logChannel, 10000)
		recordAndSend(conn, logChannel, 15000) //TODO: endSessionChannel ?!
	}

	logMessage(logChannel, "Exit Code 0")
}

func sendSong(conn net.Conn, songFileName string, endSessionChannel chan string, logChannel chan string) {
	file, err := os.Open(songFileName) // open the song that the clients wants to send to the server
	CheckError(err)
	defer func() {
		file.Close()
		logMessage(logChannel, "sendSong Done")
	}()

	buffer := make([]byte, DataFrameSize)
	// Send the song to the server (as packets)
	for {
		time.Sleep(time.Millisecond)
		tInit := time.Now().UnixMilli()
		bytesRead, err := file.Read(buffer)

		if err != nil { // When reading EOF
			logMessage(logChannel, "sendSong err:"+err.Error())
			packet := Packet{PacketType: PacketCloseChannel}
			packet.SendPacket(conn)
			break
		}
		tProcessing := time.Now().UnixMilli() - tInit
		songPacket := InitPacket(PacketRequestSong, tInit, tProcessing, bytesRead)
		songPacket.SetData(buffer)
		songPacket.SendPacket(conn)
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

func handleResponseRoutine(conn net.Conn, streamChannel chan []byte, statsChannel chan []int64, endSessionChannel chan string, logChannel chan string, waitGroup *sync.WaitGroup) {
	logMessage(logChannel, "handleResponseRoutine Start")
	defer waitGroup.Done()
	defer logMessage(logChannel, "handleResponseRoutine Done")

	for {
		var receivePacket Packet
		receivePacket.ReadPacket(conn)
		switch receivePacket.PacketType {
		case PacketRequestSong:
			// TODO: In the following instructions - why not just: streamChannel <- receivePacket.Data[]  ???
			chunk := make([]byte, receivePacket.DataSize)
			copy(chunk, receivePacket.Data[:])

			// Send the packet to the routine that pipelines the packets to mpg123
			streamChannel <- chunk // TODO: why not just: streamChannel <- receivePacket.Data[]  ???
			timeStampFinal := time.Now().UnixMilli()
			roundTripTime := timeStampFinal - int64(receivePacket.InitTime) - int64(receivePacket.ProcessingTime)
			statsChannel <- []int64{timeStampFinal, int64(receivePacket.ProcessingTime), roundTripTime}

		// todo: are the following lines necessary??
		case PacketRecord:
			streamChannel <- receivePacket.Data[:]

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
	CheckError(deleteFile(fileName))
	logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	defer logFile.Close()

	for {
		logMessage, ok := <-logChannel
		if !ok {
			// The channel has been closed
			break
		}
		fmt.Fprintln(logFile, logMessage)
	}
	fmt.Fprintf(logFile, "logRoutine Done")
}

func statsRoutine(fileName string, statsChannel chan []int64, logChannel chan string, waitGroup *sync.WaitGroup) {
	logMessage(logChannel, "statsRoutine Start")
	defer waitGroup.Done()
	defer logMessage(logChannel, "statsRoutine Done")
	var (
		roundTripTimes  []int64
		processingTimes []int64
		arrivalTimes    []int64
	)
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
			fmt.Fprintf(statisticsFile, "Packet %4d | Round Trip Time: %5d milliseconds\n", packetIndex, roundTripTime)
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
		fmt.Fprintln(statisticsFile, "Average Round Trip Time:        ", meanSendingTime, "milliseconds")
		fmt.Fprintln(statisticsFile, "Round Trip Time Jitter:         ", rttJitter, "milliseconds")
		fmt.Fprintln(statisticsFile, "Average Inter-Arrival Time:     ", meanInterArrivals, "milliseconds")
		CheckError(plotByteSlice(roundTripTimes,
			"Packets RTT Plot.png",
			"Packets RTT [milliseconds]",
			"Packet Index",
			"Packet RTT [milliseconds]"))

		CheckError(plotByteSlice(interArrivals,
			"Inter-Arrival Times.png",
			"Inter-Arrival Times [milliseconds]",
			"Packet Index",
			"Inter-Arrival Time [milliseconds]"))
	}
}

func streamRoutine(streamChannel chan []byte, logChannel chan string, waitGroup *sync.WaitGroup, workMode string) {
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

		speaker.Init(beep.SampleRate(SampleRate), AudioBufferSize)
		var buffer [][2]float64
		streamer := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
			if len(buffer) == 0 {
				chunk, ok := <-streamChannel
				if !ok {
					// The channel has been closed
					return 0, false
				}
				pcm, err := decoder.Decode(chunk, FrameSize, false)
				if err != nil {
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

func recordAndSend(conn net.Conn, logChannel chan string, durationMseconds int) {
	defer logMessage(logChannel, "sendSong Done")
	// TODO: figure how to close stream channel if needed
	//close(destinationChannel)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int16, AudioBufferSize) // Each Buffer records 20 milliseconds

	stream, err := portaudio.OpenDefaultStream(Channels, 0, SampleRate, len(in), in)
	CheckError(err)
	defer stream.Close()

	encoder, err := gopus.NewEncoder(SampleRate, Channels, gopus.Audio)
	CheckError(err)

	tInit := time.Now().UnixMilli()
	CheckError(stream.Start())

	for {
		time.Sleep(20 * time.Millisecond)
		tRecordFrame := time.Now().UnixMilli()
		CheckError(stream.Read())                                   //* Read filling the buffer by recording samples until the buffer is full
		data, err := encoder.Encode(in, FrameSize, AudioBufferSize) //* Encode PCM to Opus
		CheckError(err)
		tProcessing := time.Now().UnixMilli() - tRecordFrame
		recordPacket := InitPacket(PacketRecord, tRecordFrame, tProcessing, len(data))
		recordPacket.SetData(data)
		recordPacket.SendPacket(conn)

		select {
		case <-sig:
			CheckError(stream.Stop())
			return
		default:
			if time.Now().UnixMilli()-tInit > int64(durationMseconds) {
				CheckError(stream.Stop())
				return
			}
		}
	}
}

func play(channel chan []byte, waitGroup *sync.WaitGroup) {
	defer func() {
		waitGroup.Done()
		fmt.Println("play Done") // TODO: log to logChannel
	}()

	decoder, err := gopus.NewDecoder(SampleRate, Channels)
	CheckError(err)

	speaker.Init(beep.SampleRate(SampleRate), AudioBufferSize)
	var buffer [][2]float64
	streamer := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if len(buffer) == 0 {
			chunk, ok := <-channel
			if !ok {
				// The channel has been closed
				fmt.Println("Channel Closed")
				return 0, false
			}
			//fmt.Println("Got a chunk of size: ", len(chunk))
			pcm, err := decoder.Decode(chunk, FrameSize, false)
			if err != nil {
				fmt.Println("Error decoding chunk:", err)
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
