/*
  Portable-stripped down implementation of the PNG file format for simple verification
*/
package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
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
	Length int32
	Type   string
	Data   []byte
	CRC    uint32
}

//Verify attempts to verify the chunk with the CRC & Length of the file
func (c *Chunk) Verify() (uint32, error) {

	if int(c.Length) != len(c.Data) {
		// woop. missing bytes
		return 0, ErrorMissingBytes
	}

	dataCrc := crc32.ChecksumIEEE(c.Data)
	if dataCrc != c.CRC {
		return dataCrc, ErrorCRCMismatch
	}

	return 0, nil
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
	Chunks     map[string][]*Chunk
}

func Read(reader io.Reader) (*PNG, error) {
	buf := bufio.NewReader(reader)
	var header []byte
	var chunks = map[string][]*Chunk{}

	readingHeader := true
	lfDetection := 0

	for {
		if readingHeader {
			byteRead, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}

			if byteRead == 0x0A {
				lfDetection++
			}

			header = append(header, byteRead)

			if lfDetection >= 2 {
				readingHeader = false
			}
		} else {

			localBuf := make([]byte, 4)
			var dataLength int32
			var chunkType string
			var crc uint32

			err := binary.Read(buf, binary.BigEndian, &dataLength)

			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			_, err = buf.Read(localBuf)

			if err != nil {
				return nil, err
			}

			chunkType = string(localBuf)

			chunkData := make([]byte, dataLength)

			_, err = buf.Read(chunkData)

			if err != nil {
				return nil, err
			}

			err = binary.Read(buf, binary.BigEndian, &crc)

			if err != nil {
				return nil, err
			}

			chunk := &Chunk{
				Length: dataLength,
				Type:   chunkType,
				Data:   chunkData,
				CRC:    crc,
			}

			if _, ok := chunks[chunkType]; !ok {
				chunks[chunkType] = make([]*Chunk, 0)
			}

			v := chunks[chunkType]
			v = append(v, chunk)
			chunks[chunkType] = v
		}
	}

	return &PNG{
		FileHeader: &Header{header},
		Chunks:     chunks,
	}, nil
}
