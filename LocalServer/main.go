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
	ConnHost = ""    // ConnHost - Empty string means listen on all available interfaces
	ConnPort = "80"  // ConnPort - The port of the connection
	ConnType = "tcp" // ConnType - The type of the connection
)

func main() {
	// Enable port configuring from shell
	connPort := ConnPort
	if len(os.Args) > 1 {
		connPort = os.Args[1]
	}
	// Listen for incoming connection.
	ln, err := net.Listen(ConnType, ":"+connPort)
	if err != nil {
		log.Fatal(err)
	}
	// Close the listener when the application closes.
	defer ln.Close()
	fmt.Println("Listening on port", connPort)

	// Listen for an incoming connection.
	for {
		conn, err := ln.Accept()
		fmt.Println("opened a new connection")
		if err != nil {
			log.Fatal(err)
		}
		// Handle incoming messages
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	// Handle incoming messages
	defer conn.Close()
	for {
		message, err :=
			bufio.NewReader(conn).ReadString('\n')
		if message == "exit\n" || err == io.EOF {
			fmt.Println(conn.RemoteAddr(), "Disconnected")
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("Message Received from " + conn.RemoteAddr().String() + " " + string(message))
		newMessage := strings.ToUpper(message)
		conn.Write([]byte(newMessage))
	}
}
