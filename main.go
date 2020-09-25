package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
)

var (
	inputDirectory  = flag.String("input", "corrupted", "The path to the PNGs that need to be fixed")
	outputDirectory = flag.String("output", "fixed", "The path to the output directory")
	checkFlag       = flag.Bool("check", false, "run with this flag if you just want to check for broken PNGs")
)

var (
	ErrorMissingBytes = errors.New("missing bytes")
	ErrorCRCMismatch  = errors.New("crc mismatch")

	ErrorNotPNG              = errors.New("not a png file")
	ErrorInvalidHeaderLength = errors.New("invalid header length")
	ErrorDOSToUnixConversion = errors.New("dos to unix conversion")
	ErrorUnixToDOSConversion = errors.New("unix to dos conversion")
)

//Chunk represents a chunk within the PNG file
type Chunk struct {
	Length int
	Type   string
	Data   []byte
	CRC    uint32
}

//Verify attempts to verify the chunk with the CRC & Length of the file
func (c *Chunk) Verify() error {

	if c.Length != len(c.Data) {
		// woop. missing bytes
		return ErrorMissingBytes
	}

	if crc32.ChecksumIEEE(c.Data) != c.CRC {
		return ErrorCRCMismatch
	}

	return nil
}

type Header struct {
	HeaderBytes []byte
}

func (h *Header) Verify() error {
	if h.HeaderBytes[0] != 0x89 && string(h.HeaderBytes[1:4]) != "PNG" {
		return ErrorNotPNG
	}

	if len(h.HeaderBytes) < 8 {
		if h.HeaderBytes[5] != 0x0D {
			return ErrorDOSToUnixConversion
		}
		return ErrorInvalidHeaderLength
	}

	if h.HeaderBytes[7] != 0x0A {
		return ErrorUnixToDOSConversion
	}


	return nil
}

//PNG represents the PNG file structure
type PNG struct {
	FileHeader *Header
	Chunks     map[string]*Chunk
}

func Read(reader io.Reader) (*PNG, error) {
	buf := bufio.NewReader(reader)
	var header []byte

	readingHeader := true
	lfDetection := 0

	for {
		byteRead, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		if readingHeader {
			if byteRead == 0x1A || byteRead == 0x0A {
				lfDetection++
			}

			header = append(header, byteRead)

			if lfDetection > 2 {
				readingHeader = false
			}
		} else {
			break
		}

	}

	return &PNG{
		FileHeader: &Header{header},
	}, nil
}

func main() {
	b, err := ioutil.ReadFile("speedsilver.png")

	if err != nil {
		log.Fatalln(err)
	}

	pngFile, err := Read(bytes.NewReader(b))

	if err != nil {
		log.Fatalln(err)
	}

	verifyError := pngFile.FileHeader.Verify()

	if verifyError != nil {
		log.Println(verifyError.Error())
	}

	fmt.Printf("length: %d, header: %s",  len(pngFile.FileHeader.HeaderBytes), hex.Dump(pngFile.FileHeader.HeaderBytes))

}
