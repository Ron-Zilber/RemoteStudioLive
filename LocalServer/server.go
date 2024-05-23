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

// Server type
type Server struct {
	connSpecs ConnSpecs
}

func (server *Server) start() {
	specs := InitConnSpecs(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	server.connSpecs = *specs
	ln, err := net.Listen(server.connSpecs.Type, ":"+server.connSpecs.Port)
	CheckError(err)
	defer ln.Close()
	fmt.Println("Listening on " + server.connSpecs.IP + ":" + server.connSpecs.Port)

	// Listen for an incoming connection.
	for {
		conn, err := ln.Accept()
		CheckError(err)
		fmt.Println("Connected to:", conn.RemoteAddr().String())

		// Handle incoming messages
		go handleConnection(conn, server.connSpecs.OpMode)
	}
}

func main() {
	server := &Server{}
	server.start()
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
