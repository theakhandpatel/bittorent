package main

import (
	"encoding/json"
	"fmt"
	"os"

	torrentMeta "github.com/codecrafters-io/bittorrent-starter-go/internal/TorrentMeta"
	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decodeHandler(bencodedValue string) {
	decoded, err := bencode.Decode(bencodedValue)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonOutput, _ := json.Marshal(decoded)
	if string(jsonOutput) != "null" {
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("[]")
	}
}

func infoHandler(torrentFilePath string) {
	torrent, err := torrentMeta.NewFromFile(torrentFilePath)
	if err != nil {
		fmt.Println("Error parsing file")
		return
	}
	fmt.Printf("Tracker URL: %s\nLength: %d\n", torrent.Announce, torrent.Info.Length)
	fmt.Printf("Info Hash: %s\n", torrent.InfoHash)

	pieces, err := torrent.GetPieces()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Piece Hashes:")
	for _, piece := range pieces {
		fmt.Println(piece)
	}

}

func main() {
	command := os.Args[1]

	if command == "decode" {
		decodeHandler(os.Args[2])
	} else if command == "info" {
		infoHandler(os.Args[2])
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
