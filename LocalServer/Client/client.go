// Client src code
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"log"
)

const (
	ServerIP   = "172.23.175.237" // ServerIP   - Replace with the actual IP address of your server
	ServerPort = "8080"             // ServerPort - The port number of the server
	ConnType   = "tcp"            // ConnType   - The type of the connection
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
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		//sendDefaultMessage(conn, i)
		if !sendMessage(conn) {
			break
		}
		handleRequest(conn)
	}

}

func sendMessage(conn net.Conn) (status bool) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Text to send: ")
	text, _ := reader.ReadString('\n')
	if text == "exit\n" {
		fmt.Println(conn.LocalAddr().String(), "Disconnecting")
		fmt.Fprintf(conn, text+"\n")
		return false

	}
	// Send the text to the server:
	fmt.Fprintf(conn, text+"\n")
	return true
}

func handleRequest(conn net.Conn) {
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message from the server: " + message)
}
