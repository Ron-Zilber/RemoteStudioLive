// Client src code
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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
	text, _ := reader.ReadString('\n')
	if text == "exit\n" {
		fmt.Println(conn.LocalAddr().String(), "Disconnecting")
		fmt.Fprintf(conn, text+"\n")
		return false

	} else if text == "lorem\n" {
		sendFile(conn, "lorem ipsum.txt")
		return true
	}
	// Send the text to the server:
	fmt.Fprintf(conn, text+"\n")
	return true
}

func handleRequest(conn net.Conn) {
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message from the server: " + message)
}

func sendFile(conn net.Conn, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
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
