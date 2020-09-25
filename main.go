package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
)

var (
	inputDirectory  = flag.String("input", "corrupted", "The path to the PNGs that need to be fixed")
	outputDirectory = flag.String("output", "fixed", "The path to the output directory")
	checkFlag       = flag.Bool("check", false, "run with this flag if you just want to check for broken PNGs")
)

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

	for _, v := range pngFile.Chunks {
		for _, chunk := range v {
			if chunk.Type == "IEND" {
				break
			}
			crc, verifiedError := chunk.Verify()

			if verifiedError == nil {
				log.Printf("Chunk %s is OK!\n", chunk.Type)
			} else {
				log.Printf("Chunk %s has an error! Chunk CRC: %x, Data CRC: %x Error %s\n", chunk.Type, chunk.CRC, crc, verifiedError.Error())
			}
		}
	}
}
