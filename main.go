package main

import (
	"log"
	"os"

	"github.com/Kud1nov/torrent-client/torrentfile"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentfile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", tf.AnnounceList)
	os.Exit(0)
	//log.Printf("%+v", tf)

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Printf("%s", "Download Error")
		log.Fatal(err)
	}
}
