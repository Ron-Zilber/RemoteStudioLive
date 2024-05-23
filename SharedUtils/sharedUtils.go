// Package sharedutils contains shared functions and constants
package sharedutils

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	//BufferSize is the size of a buffer
	BufferSize    = bufio.MaxScanTokenSize / 64 // BufferSize - The size of the packets when transmitting a song
	DataFrameSize = BufferSize - 24             // DataFrameSize - The max size of the data part in a packet
)

// Packet is the definition for a packet in the module
type Packet struct {
	PacketType     uint32
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

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type commonChunk struct {
	NumChans      int16
	NumSamples    int32
	BitsPerSample int16
	SampleRate    [10]byte
}

type ID [4]byte

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
		packet.InitTime = binary.LittleEndian.Uint64(buf[4:12])
		packet.ProcessingTime = binary.LittleEndian.Uint64(buf[12:20])
		packet.DataSize = binary.LittleEndian.Uint32(buf[20:24])
		if packet.DataSize != 0 {
			copy(packet.Data[0:DataFrameSize], buf[24:BufferSize])
		}
	}

	return packetRead
}

// SendPacket encodes a packet into a binary byte slice and send it through a link
func (packet *Packet) SendPacket(conn net.Conn) {
	buf := make([]byte, BufferSize)
	binary.LittleEndian.PutUint32(buf[0:], packet.PacketType)
	binary.LittleEndian.PutUint64(buf[4:], uint64(packet.InitTime))
	binary.LittleEndian.PutUint64(buf[12:], uint64(packet.ProcessingTime))
	binary.LittleEndian.PutUint32(buf[20:], packet.DataSize)
	copy(buf[24:BufferSize], packet.Data[0:packet.DataSize])
	_, err := conn.Write(buf)
	CheckError(err)
}

// InitPacket initializing a packet
func InitPacket(packetType int, initTime int64, processingTime int64, dataSize int) *Packet {
	return &Packet{
		PacketType:     uint32(packetType),
		InitTime:       uint64(initTime),
		ProcessingTime: uint64(processingTime),
		DataSize:       uint32(dataSize),
	}
}

// SetData sets the data for a packet
func (packet *Packet) SetData(dataBuffer []byte) {
	copy(packet.Data[:], dataBuffer)
}

func record(fileName string, duration int64) {

	fmt.Println("Recording.  Press Ctrl-C to stop.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	if !strings.HasSuffix(fileName, ".aiff") {
		fileName += ".aiff"
	}
	f, err := os.Create(fileName)
	CheckError(err)

	// form chunk
	_, err = f.WriteString("FORM")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //total bytes
	_, err = f.WriteString("AIFF")
	CheckError(err)

	// common chunk
	_, err = f.WriteString("COMM")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(18)))           //size
	CheckError(binary.Write(f, binary.BigEndian, int16(1)))            //channels
	CheckError(binary.Write(f, binary.BigEndian, int32(0)))            //number of samples
	CheckError(binary.Write(f, binary.BigEndian, int16(32)))           //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	CheckError(err)

	// sound chunk
	_, err = f.WriteString("SSND")
	CheckError(err)
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //size
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //offset
	CheckError(binary.Write(f, binary.BigEndian, int32(0))) //block
	nSamples := 0
	defer func() {
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		_, err = f.Seek(4, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(totalBytes)))
		_, err = f.Seek(22, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(nSamples)))
		_, err = f.Seek(42, 0)
		CheckError(err)
		CheckError(binary.Write(f, binary.BigEndian, int32(4*nSamples+8)))
		CheckError(f.Close())
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	CheckError(err)
	defer stream.Close()

	CheckError(stream.Start())

	tStart := time.Now().Unix()
	for {
		CheckError(stream.Read())
		CheckError(binary.Write(f, binary.BigEndian, in))
		nSamples += len(in)
		select {
		case <-sig:
			CheckError(stream.Stop())
			return

		default:

			if time.Now().Unix()-tStart > duration {
				CheckError(stream.Stop())
				return
			}
		}

	}
	fmt.Println("Stopping the stream")
	CheckError(stream.Stop())
}

func play(fileName string) {
	fmt.Println("Playing.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	f, err := os.Open(fileName)
	CheckError(err)
	defer f.Close()

	id, data, err := readChunk(f)
	CheckError(err)
	if id.String() != "FORM" {
		fmt.Println("bad file format")
		return
	}
	_, err = data.Read(id[:])
	CheckError(err)
	if id.String() != "AIFF" {
		fmt.Println("bad file format")
		return
	}
	var c commonChunk
	var audio io.Reader
	for {
		id, chunk, err := readChunk(data)
		if err == io.EOF {
			break
		}
		CheckError(err)
		switch id.String() {
		case "COMM":
			CheckError(binary.Read(chunk, binary.BigEndian, &c))
		case "SSND":
			chunk.Seek(8, 1) //ignore offset and block
			audio = chunk
		default:
			fmt.Printf("ignoring unknown chunk '%s'\n", id)
		}
	}

	//assume 44100 sample rate, mono, 32 bit

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int32, 8192)
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), &out)
	CheckError(err)
	defer stream.Close()

	CheckError(stream.Start())
	defer stream.Stop()
	for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
		if len(out) > remaining {
			out = out[:remaining]
		}
		err := binary.Read(audio, binary.BigEndian, out)
		if err == io.EOF {
			break
		}
		CheckError(err)
		CheckError(stream.Write())
		select {
		case <-sig:
			return
		default:
		}
	}
}

func readChunk(r readerAtSeeker) (id ID, data *io.SectionReader, err error) {
	_, err = r.Read(id[:])
	if err != nil {
		return
	}
	var n int32
	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}
	off, _ := r.Seek(0, 1)
	data = io.NewSectionReader(r, off, int64(n))
	_, err = r.Seek(int64(n), 1)
	return
}

func (id ID) String() string {
	return string(id[:])
}