// Package server is a nice package
package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	CONN_HOST = "" // Empty string means listen on all available interfaces
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func main() {
	// Listen for incoming connection.
	l, err := net.Listen(CONN_TYPE, ":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}

	name, err := os.Hostname()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close() // defer pushes the call to Close() to the stack s.t it will be executed before the server() function returning
	host, _ := net.LookupHost(name)
	fmt.Println("Listening on "  + host[0] + ":",CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		// logs an incoming message

		fmt.Printf("Received message %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine
		go handleRequest(conn)
	}
}

// Handles incoming requests
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data
	buff := make([]byte, 1024)
	// Read the incoming connection into the buffer
	reqLen, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// Builds the message
	message := "Hi, I received your message!\nIt was "
	message += strconv.Itoa(reqLen)
	message += " bytes long and that's what it said:\n "
	n := bytes.Index(buff, []byte{0})
	message += string(buff[:n-1])
	message += "\n"
	
	// Write the message in the connection channel
	conn.Write([]byte(message))
	// Close the connection when you're done with it
	conn.Close()
}
