// Package shared contains shared functions and constants
package shared

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const (
	//BufferSize is the size of a buffer
	BufferSize = bufio.MaxScanTokenSize / 64 // BufferSize - The size of the packets when transmitting a song
)

// Packet is the definition for a packet in the module
type Packet struct {
	PacketType     uint32
	PacketInitTime uint64
	DataSize       uint32
	PacketData     [BufferSize - 16]byte
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ReadPacket(conn net.Conn, packet *Packet) bool {
	buf := make([]byte, BufferSize)
	packetRead := false

	for {
		packetLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("packet read err")
			break
		}

		if packetLen == BufferSize {
			packetRead = true
			break
		}
	}

	if packetRead {
		packet.PacketType = binary.LittleEndian.Uint32(buf[0:4])
		packet.PacketInitTime = binary.LittleEndian.Uint64(buf[4:12])
		packet.DataSize = binary.LittleEndian.Uint32(buf[12:16])
		if packet.DataSize != 0 {
			copy(packet.PacketData[0:BufferSize-16], buf[16:BufferSize])
		}
	}

	return packetRead
}

func SendPacket(conn net.Conn, packet *Packet) {
	buf := make([]byte, BufferSize)
	binary.LittleEndian.PutUint32(buf[0:], packet.PacketType)
	binary.LittleEndian.PutUint64(buf[4:], uint64(packet.PacketInitTime))
	binary.LittleEndian.PutUint32(buf[12:], packet.DataSize)
	copy(buf[16:BufferSize], packet.PacketData[0:packet.DataSize])
	conn.Write(buf)
}
