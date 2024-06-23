// Package sharedutils contains shared functions and constants
package sharedutils

import (
	"bufio"
	"encoding/binary"
	"log"
	"net"
)

const (
	//BufferSize is the size of a buffer
	BufferSize    = bufio.MaxScanTokenSize / 64 // BufferSize - The size of the packets when transmitting a song
	DataFrameSize = BufferSize - 28             // DataFrameSize - The max size of the data part in a packet
)

// Packet is the definition for a packet in the module
type Packet struct {
	PacketType     uint32
	SerialNumber   uint32
	InitTime       uint64
	ProcessingTime uint64
	DataSize       uint32
	Data           [DataFrameSize]byte
}

// ConnSpecs structs the specifications for the connection
type ConnSpecs struct {
	Type   string
	IP     string
	Port   string
	OpMode string
}

// InitConnSpecs constructs Connection
func InitConnSpecs(connType string, connIP string, connPort string, opMode string) *ConnSpecs {
	return &ConnSpecs{
		Type:   connType,
		IP:     connIP,
		Port:   connPort,
		OpMode: opMode,
	}
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ReadPacket accepts a binary byte slice from the link and decode it into a packet structure
func (packet *Packet) ReadPacket(conn net.Conn) bool {
	buf := make([]byte, BufferSize)
	packetRead := false

	for {
		packetLen, err := conn.Read(buf)
		CheckError(err)

		if packetLen == BufferSize {
			packetRead = true
			break
		}
	}

	if packetRead {
		packet.PacketType = binary.LittleEndian.Uint32(buf[0:4])
		packet.SerialNumber = binary.LittleEndian.Uint32(buf[4:8])
		packet.InitTime = binary.LittleEndian.Uint64(buf[8:16])
		packet.ProcessingTime = binary.LittleEndian.Uint64(buf[16:24])
		packet.DataSize = binary.LittleEndian.Uint32(buf[24:28])
		if packet.DataSize != 0 {
			copy(packet.Data[0:DataFrameSize], buf[28:BufferSize])
		}
	}
	return packetRead
}

// SendPacket encodes a packet into a binary byte slice and send it through a link
func (packet *Packet) SendPacket(conn net.Conn) {
	buf := make([]byte, BufferSize)
	binary.LittleEndian.PutUint32(buf[0:], packet.PacketType)
	binary.LittleEndian.PutUint32(buf[4:], packet.SerialNumber)
	binary.LittleEndian.PutUint64(buf[8:], packet.InitTime)
	binary.LittleEndian.PutUint64(buf[16:], packet.ProcessingTime)
	binary.LittleEndian.PutUint32(buf[24:], packet.DataSize)
	copy(buf[28:BufferSize], packet.Data[0:packet.DataSize])
	_, err := conn.Write(buf)
	CheckError(err)
}

// InitPacket initializing a packet
func InitPacket(packetType int, serialNumber int, initTime int64, processingTime int64, dataSize int) *Packet {
	return &Packet{
		PacketType:     uint32(packetType),
		SerialNumber:   uint32(serialNumber),
		InitTime:       uint64(initTime),
		ProcessingTime: uint64(processingTime),
		DataSize:       uint32(dataSize),
	}
}

// SetData sets the data for a packet
func (packet *Packet) SetData(dataBuffer []byte) {
	copy(packet.Data[:], dataBuffer)
}

// GetPacketSize returns the size of the packet in bytes
func (packet *Packet) GetPacketSize() int {
	return len(packet.Data) + 28 // 28 is the summarized length of all other fields of the packet
}
