package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
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
		fmt.Println(peer.String())
	}
}

func handshakeHandler(torrentFilePath string, ipPort string) {
	torrent, err := torrentMeta.NewFromFile(torrentFilePath)
	if err != nil {
		fmt.Println("Error parsing file")
		return
	}

	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		fmt.Println(err)
	}

	recieverID, err := torrent.Handshake(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Peer ID: %s", recieverID)
}

// func downloadPieceHandler(outputPath string, torrentFilePath string, pieceNum int) {
// 	torrent, err := torrentMeta.NewFromFile(torrentFilePath)
// 	if err != nil {
// 		fmt.Println("Error parsing file")
// 		return
// 	}
// 	peers, err := torrent.DiscoverPeers()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	peer := peers[0]
// 	conn, err := net.Dial("tcp", peer.String())
// 	if err != nil {
// 		fmt.Println("Error connecting to peer:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	receiverID, err := torrent.Handshake(conn)
// 	if err != nil {
// 		fmt.Println("Error during handshake:", err)
// 		return
// 	}
// 	fmt.Printf("Peer ID: %s\n", receiverID)

// 	_, err = torrentMeta.WaitFor(torrentMeta.MSG_BITFIELD, conn)
// 	if err != nil {
// 		fmt.Println("Error reading bitfield message:", err)
// 		return
// 	}

// 	sendMsg := torrentMeta.NewPeerMessage(torrentMeta.MSG_INTERESTED, nil)
// 	err = torrentMeta.WritePeerMessage(conn, sendMsg)
// 	if err != nil {
// 		fmt.Println("Error sending Interested message:", err)
// 		return
// 	}

// 	_, err = torrentMeta.WaitFor(torrentMeta.MSG_UNCHOKE, conn)
// 	if err != nil {
// 		fmt.Println("Error reading Unchoke message:", err)
// 		return
// 	}

// }

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
	} else if command == "download_piece" {

	} else {
		pieceNum := 0 //strconv.Atoi(os.Args[4])
		downloadPieceHandler("./data/", os.Args[1], pieceNum)
		// fmt.Println("Unknown command: " + command)
		// os.Exit(1)
	}
}
