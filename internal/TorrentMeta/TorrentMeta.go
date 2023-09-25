package torrentMeta

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
)

type TorrentMeta struct {
	Announce string
	Info     struct {
		Length       int
		Name         string
		Piece_length int
		Pieces       string
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
			Length       int
			Name         string
			Piece_length int
			Pieces       string
		}{
			Length:       metaInfoDict["length"].(int),
			Name:         metaInfoDict["name"].(string),
			Piece_length: metaInfoDict["piece length"].(int),
			Pieces:       metaInfoDict["pieces"].(string),
		},
		InfoHash: infoHash,
	}, nil
}
