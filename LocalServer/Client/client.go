package main

import (
	"fmt"
	"net"
	"os"
)

const (
	SERVER_IP   = "192.168.1.2" // Replace with the actual IP address of your server
	SERVER_PORT = "3333"
	CONN_TYPE   = "tcp"
)

func main() {
	// Connect to the server
	conn, err := net.Dial(CONN_TYPE, SERVER_IP+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Send a message to the server
	message := "Hello from the client!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err.Error())
		os.Exit(1)
	}

	// Read the response from the server
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading response:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Server response:", string(buffer))
}
