// Packet main for playground
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	//"github.com/hraban/opus"
	//"github.com/hraban/opus"
	//"gopkg.in/hraban/opus.v2"
)

const (
	SampleRate = 48000 // SampleRate is the number of bits used to represent a full second of audio sampling
	Channels   = 2     // Channels - 1 for mono; 2 for stereo
	FrameSize  = 960   // FrameSize of 960 gives 20 ms (for 48kHz sampling) which is the Opus recommendation

	BufferSize = FrameSize * Channels // BufferSize let the buffer hold multiple frames
)

func main() {

	var num1 int64 = time.Now().UnixMicro()
	
	var num2 uint64 = uint64(num1)
	num3 := int64(num2)
	fmt.Println(num1, "\t",num2, "\t", num3)
}

func printFile(fileName string) {
	f, err := os.Open(fileName)
	CheckError(err)

	_, err = io.Copy(os.Stdout, f)
	CheckError(err)
	CheckError(f.Close())
}

func deleteFile(fileName string) error {
	// Check if the file exists
	if _, err := os.Stat(fileName); err == nil {
		// File exists, delete it

		err := os.Remove(fileName)
		if err != nil {
			return err
		}
	} else if !strings.HasSuffix(err.Error(), "no such file or directory") {
		// Some other error occurred
		return err
	}
	return nil
}

// CheckError General error handling
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}
