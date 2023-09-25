package main

import (
	"encoding/hex"
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
	fmt.Printf("Info Hash: %s\n", hex.EncodeToString(torrent.InfoHash))
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
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

func peersHandler(torrentFilePath string) {
	torrent, err := torrentMeta.NewFromFile(torrentFilePath)
	if err != nil {
		fmt.Println("Error parsing file")
		return
	}
	peers, err := torrent.DiscoverPeers()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, peer := range peers {
		fmt.Printf("%s:%d\n", peer.IP, peer.Port)
	}
}

func handshakeHandler(torrentFilePath string, peerString string) {
	torrent, err := torrentMeta.NewFromFile(torrentFilePath)
	if err != nil {
		fmt.Println("Error parsing file")
		return
	}
	recieverID, err := torrent.Handshake(peerString)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Peer ID: %s", recieverID)
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		decodeHandler(os.Args[2])
	} else if command == "info" {
		infoHandler(os.Args[2])
	} else if command == "peers" {
		peersHandler(os.Args[2])
	} else if command == "handshake" {
		handshakeHandler(os.Args[2], os.Args[3])
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
