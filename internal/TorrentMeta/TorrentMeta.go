package torrentMeta

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
)

const (
	ProtocolIdentifier = "BitTorrent protocol"
	ReservedBytes      = "\x00\x00\x00\x00\x00\x00\x00\x00"
)

type TorrentMeta struct {
	Announce string
	Info     struct {
		Length      int
		Name        string
		PieceLength int
		PiecesHash  string
	}
	InfoHash   []byte
	PeerID     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
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
	infoHash := (shaHash[:])

	tm := TorrentMeta{}
	tm.PeerID = generatePeerID()
	tm.Info.Length = metaInfoDict["length"].(int)
	tm.Info.Name = metaInfoDict["name"].(string)
	tm.Info.PieceLength = metaInfoDict["piece length"].(int)
	tm.Info.PiecesHash = metaInfoDict["pieces"].(string)
	tm.InfoHash = infoHash
	tm.Announce = metaDict["announce"].(string)
	tm.Port = 6881
	tm.Uploaded = 0
	tm.Downloaded = 0
	tm.Left = tm.Info.Length

	return tm, nil
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

// ConstructTrackerURL constructs the tracker URL with required query parameters.
func (tm *TorrentMeta) ConstructTrackerURL() (string, error) {
	// Create a URL object
	trackerURL, err := url.Parse(tm.Announce)
	if err != nil {
		return "", err
	}

	// Prepare query parameters
	query := trackerURL.Query()
	query.Add("info_hash", string(tm.InfoHash))
	query.Add("peer_id", tm.PeerID)
	query.Add("port", fmt.Sprint(tm.Port))             // Set your desired port here
	query.Add("uploaded", fmt.Sprint(tm.Uploaded))     // Total uploaded bytes
	query.Add("downloaded", fmt.Sprint(tm.Downloaded)) // Total downloaded bytes
	query.Add("left", fmt.Sprint(tm.Left))             // Number of bytes left to download
	query.Add("compact", "1")                          // Use compact representation

	// Set the query parameters
	trackerURL.RawQuery = query.Encode()

	return trackerURL.String(), nil
}

func (tm *TorrentMeta) DiscoverPeers() ([]*Peer, error) {
	trackerURL, err := tm.ConstructTrackerURL()
	if err != nil {
		return nil, err
	}
	response, err := http.Get(trackerURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	trackerResponse, err := bencode.Decode(string(body))
	trackerResponseMap := trackerResponse.(map[string]interface{})
	if err != nil {
		return nil, err
	}

	peersStr, ok := trackerResponseMap["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("peers not recieved")
	}

	var peerList []*Peer
	if len(peersStr)%6 != 0 {
		return nil, fmt.Errorf("invalid peers string length")
	}

	for i := 0; i < len(peersStr); i += 6 {
		ipBytes := peersStr[i : i+4]
		portBytes := peersStr[i+4 : i+6]

		ip := net.IP(ipBytes).String()
		port := int(binary.BigEndian.Uint16([]byte(portBytes)))

		peerList = append(peerList, &Peer{IP: ip, Port: port})
	}

	return peerList, nil
}

func (torrent *TorrentMeta) Handshake(ipPort string) (string, error) {
	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Construct the handshake message
	handshakeMessage := []byte{byte(len(ProtocolIdentifier))}
	handshakeMessage = append(handshakeMessage, []byte(ProtocolIdentifier)...)
	handshakeMessage = append(handshakeMessage, ReservedBytes...)
	handshakeMessage = append(handshakeMessage, []byte(torrent.InfoHash)...)
	handshakeMessage = append(handshakeMessage, []byte(torrent.PeerID)...)

	// Send the handshake message
	_, err = conn.Write(handshakeMessage)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Receive and parse the handshake response
	response := make([]byte, 68) // A handshake response is always 68 bytes
	_, err = conn.Read(response)
	if err != nil {
		return "", err
	}
	peerID := fmt.Sprintf("%x", response[48:])

	return peerID, nil
}

type Peer struct {
	IP   string
	Port int
}

func generatePeerID() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	// rand.Seed(time.Now().UnixNano())

	b := make([]rune, 20)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
