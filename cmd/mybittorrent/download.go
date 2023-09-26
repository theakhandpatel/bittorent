package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"

	torrentMeta "github.com/codecrafters-io/bittorrent-starter-go/internal/TorrentMeta"
)

func downloadPieceHandler(outputPath string, torrentFilePath string, pieceNum int) {
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
	peer := peers[0]
	conn, err := net.Dial("tcp", peer.String())
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	receiverID, err := torrent.Handshake(conn)
	if err != nil {
		fmt.Println("Error during handshake:", err)
		return
	}
	fmt.Printf("Peer ID: %s\n", receiverID)

	respMsg, err := torrentMeta.WaitFor(torrentMeta.MSG_BITFIELD, conn)
	if err != nil {
		fmt.Println("Error reading bitfield message:", err)
		return
	}

	fmt.Printf("Received bitfield peer message: Length=%d, ID=%d, Payload=%s\n", respMsg.Length, respMsg.ID, respMsg.Payload)

	sendMsg := torrentMeta.NewPeerMessage(torrentMeta.MSG_INTERESTED, nil)
	err = torrentMeta.WritePeerMessage(conn, sendMsg)
	if err != nil {
		fmt.Println("Error sending Interested message:", err)
		return
	}

	respMsg, err = torrentMeta.WaitFor(torrentMeta.MSG_UNCHOKE, conn)
	if err != nil {
		fmt.Println("Error reading Unchoke message:", err)
		return
	}
	fmt.Printf("Received Unchoke peer message: Length=%d, ID=%d, Payload=%s\n", respMsg.Length, respMsg.ID, respMsg.Payload)

	// Download piece 0 in blocks
	pieceLength := torrent.Info.PieceLength

	fmt.Println(pieceLength, " is the size of piece")
	fmt.Println(torrentMeta.DefaultBlockSize, " is the size of block")

	startOffset := pieceNum * pieceLength

	numBlocks := (pieceLength + torrentMeta.DefaultBlockSize - 1) / torrentMeta.DefaultBlockSize
	fmt.Println(numBlocks, " is the num of blocks")
	pieceData := make([]byte, pieceLength)

	for i := 0; i < numBlocks; i++ {
		blockStart := startOffset + i*torrentMeta.DefaultBlockSize
		blockEnd := blockStart + torrentMeta.DefaultBlockSize
		if blockEnd > startOffset+pieceLength {
			blockEnd = startOffset + pieceLength
		}

		blockData, err := downloadBlock(conn, blockStart, blockEnd)
		if err != nil {
			fmt.Println("Error downloading block:", err)
			return
		}
		fmt.Println("Downloaded Block", i)
		// Copy the block into the appropriate position in the pieceData
		copy(pieceData[blockStart-startOffset:], blockData)
	}

	if torrent.IsAuthentic(pieceData, 0) {
		fmt.Println("couldn't downlaod")
		return
	}
	// Save the complete piece to a file
	pieceFilePath := filepath.Join(outputPath, "0")
	err = savePieceToFile(pieceData, pieceFilePath)
	if err != nil {
		fmt.Println("Error saving piece to file:", err)
		return
	}

	fmt.Printf("Piece %d downloaded to %s\n", pieceNum, outputPath)
}

func savePieceToFile(pieceData []byte, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(pieceData)
	if err != nil {
		return err
	}

	return nil
}

func downloadBlock(conn net.Conn, startOffset, endOffset int) ([]byte, error) {

	requestMsg := createRequestMessage(startOffset, endOffset)

	// Send the request message to the peer
	err := torrentMeta.WritePeerMessage(conn, requestMsg)
	if err != nil {
		return nil, err
	}

	// Wait for the piece message from the peer
	pieceMsg, err := torrentMeta.WaitFor(torrentMeta.MSG_PIECE, conn)
	if err != nil {
		return nil, err
	}

	return pieceMsg.Payload, nil
}

func createRequestMessage(startOffset, endOffset int) *torrentMeta.PeerMessage {
	index := startOffset / torrentMeta.DefaultBlockSize
	begin := startOffset % torrentMeta.DefaultBlockSize
	length := endOffset - startOffset

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return torrentMeta.NewPeerMessage(torrentMeta.MSG_REQUEST, payload)
}
