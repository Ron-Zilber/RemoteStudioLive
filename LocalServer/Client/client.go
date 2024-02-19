// Client src code
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ServerIP   = "172.23.175.237" // ServerIP   - Replace with the actual IP address of your server
	ServerPort = "8080"           // ServerPort - The port number of the server
	ConnType   = "tcp"            // ConnType   - The type of the connection
)

func main() {
	// Enable server IP and port configuring from shell
	serverPort, serverIP := ServerPort, ServerIP

	if len(os.Args) > 2 {
		serverIP = os.Args[1]
		serverPort = os.Args[2]
	}
	// Connect to the server
	conn, err := net.Dial(ConnType, serverIP+":"+serverPort)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		if !sendMessage(conn) {
			break
		}
		handleRequest(conn)
	}
}

func sendMessage(conn net.Conn) (status bool) {
	reader := bufio.NewReader(os.Stdin)
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
	return true
}

func handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
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

func sendFile(conn net.Conn, fileName string) {
	file, err := os.Open(fileName)
	checkError(err)
	// close file on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	// make a buffer to keep chunks that are read
	reader := bufio.NewReader(file)
	for {
		// read a chunk
		s, err := reader.ReadString('\n')
		for err == nil {
			// send a chunk
			fmt.Fprintf(conn, s+"\n")
			handleRequest(conn)
			s, err = reader.ReadString('\n')
		}
		// Send the last chunk
		fmt.Fprintf(conn, s+"\n")
		break
	}
}

// checkError General error handling
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func pingServer(conn net.Conn) {
	i := 1
	count := 100
	frequency := 100 * time.Millisecond
	for i < count {
		fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n")
		handleRequest(conn)
		time.Sleep(frequency)
		i++
	}
	fmt.Fprintf(conn, "Ping "+strconv.Itoa(i)+"\n") // Send the last ping
}
