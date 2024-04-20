// Server src code
package main

import (
	. "RemoteStudioLive/SharedUtils"
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	ConnHost = ""        // ConnHost - Empty string means listen on all available interfaces
	ConnPort = "8080"    // ConnPort - The port of the connection
	ConnType = "tcp"     // ConnType - The type of the connection
	OpMode   = "default" // OpMode - The operation mode
)

func main() {

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
				fmt.Println(conn.RemoteAddr(), "Disconnected")
				break
			}
			// Send chunk back to the client
			_, writeErr := conn.Write(buffer[:bytesRead])
			CheckError(writeErr)
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
