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

func main() {

	server := &Server{}
	specs := InitConnSpecs(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	server.connSpecs = *specs

	server.start()
}

func (server *Server) start() {
	switch server.connSpecs.Type {
	case "tcp":
		server.startTCP()

	case "udp":
		server.startUDP()

	default:
		fmt.Println("Wrong arguments for server initialization")
	}
}

func (server *Server) startTCP() {
	ln, err := net.Listen(server.connSpecs.Type, ":"+server.connSpecs.Port)
	CheckError(err)
	defer ln.Close()
	fmt.Println("Listening tcp on " + server.connSpecs.IP + ":" + server.connSpecs.Port)

	// Listen for an incoming connection.
	for {
		conn, err := ln.Accept()
		CheckError(err)
		fmt.Println("Connected to:", conn.RemoteAddr().String())

		// Handle incoming messages
		go handleConnection(conn, server.connSpecs.OpMode)
	}
}

func (server *Server) startUDP() {
	ln, err := net.ListenPacket("udp", ":"+server.connSpecs.Port)
	CheckError(err)
	defer ln.Close()
	fmt.Println("Listening udp on " + server.connSpecs.IP + ":" + server.connSpecs.Port)

	for {
		buffer := make([]byte, bufio.MaxScanTokenSize)
		bytesRead, address, err := ln.ReadFrom(buffer)
		CheckError(err)
		fmt.Println("Received packet from: ", address)
		go func(b []byte, a net.Addr) {
			_, err = ln.WriteTo(buffer[:bytesRead], address) // Send chunk back to the client
			CheckError(err)
		}(buffer, address)
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
