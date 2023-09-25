package torrentMeta

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
)

type TorrentMeta struct {
	Announce string
	Info     struct {
		Length      int
		Name        string
		PieceLength int
		PiecesHash  string
	}
	InfoHash string
}

func NewFromFile(filePath string) (TorrentMeta, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TorrentMeta{}, err
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return TorrentMeta{}, err
	}
	decodedContent, err := bencode.Decode(string(fileContent))
	if err != nil {
		return TorrentMeta{}, err
	}

	metaDict := decodedContent.(map[string]interface{})
	metaInfoDict := metaDict["info"].(map[string]interface{})
	encodedInfo, err := bencode.Encode(metaInfoDict)
	if err != nil {
		return TorrentMeta{}, err
	}
	shaHash := sha1.Sum([]byte(encodedInfo))
	infoHash := hex.EncodeToString(shaHash[:])

	return TorrentMeta{
		Announce: metaDict["announce"].(string),
		Info: struct {
			Length      int
			Name        string
			PieceLength int
			PiecesHash  string
		}{
			Length:      metaInfoDict["length"].(int),
			Name:        metaInfoDict["name"].(string),
			PieceLength: metaInfoDict["piece length"].(int),
			PiecesHash:  metaInfoDict["pieces"].(string),
		},
		InfoHash: infoHash,
	}, nil
}

func (tm *TorrentMeta) GetPieces() ([]string, error) {

	if len(tm.Info.PiecesHash)%20 != 0 {
		return nil, fmt.Errorf("pieceHashes length is not a multiple of 20 bytes")
	}

	numPieces := len(tm.Info.PiecesHash) / 20
	hashes := make([]string, numPieces)

	for i := 0; i < numPieces; i++ {
		startIndex := i * 20
		endIndex := startIndex + 20
		pieceHash := tm.Info.PiecesHash[startIndex:endIndex]
		hashes[i] = hex.EncodeToString([]byte(pieceHash))

	}
	return hashes, nil
}
