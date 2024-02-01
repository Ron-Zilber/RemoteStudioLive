// Client src code
package main

import (
	"fmt"
	"net"
	"os"
)

const (
	ServerIP   = "172.23.175.237" // ServerIP - Replace with the actual IP address of your server
	ServerPort = "8080"           // ServerPort - The port number of the server
	ConnType   = "tcp"            // ConnType - The type of the connection
)

func main() {
	// Enable port configuring from shell
	serverPort := ServerPort
	if len(os.Args) > 1 {
		serverPort = os.Args[1]
	}
	// Connect to the server
	conn, err := net.Dial(ConnType, ServerIP+":"+serverPort)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Send a message to the server
	message := "Hello from the client! "
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
