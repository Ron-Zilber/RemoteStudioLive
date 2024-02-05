// Server src code
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	ConnHost = ""     // ConnHost - Empty string means listen on all available interfaces
	ConnPort = "8080" // ConnPort - The port of the connection
	ConnType = "tcp"  // ConnType - The type of the connection
)

func main() {
	// Enable port configuring from shell
	connPort := ConnPort
	if len(os.Args) > 1 {
		connPort = os.Args[1]
	}
	// Listen for incoming connection.
	ln, err := net.Listen(ConnType, ":"+connPort)
	CheckError(err)
	// Close the listener when the application closes.
	defer ln.Close()
	fmt.Println("Listening on port", connPort)

	// Listen for an incoming connection.
	for {
		conn, err := ln.Accept()
		fmt.Println("opened a new connection")
		CheckError(err)
		// Handle incoming messages
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	// Handle incoming messages
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		message, err :=
			reader.ReadString('\n')
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

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
