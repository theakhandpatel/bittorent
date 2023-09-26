package torrentMeta

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	MSG_CHOKE         = 0
	MSG_UNCHOKE       = 1
	MSG_INTERESTED    = 2
	MSG_NOTINTERESTED = 3
	MSG_HAVE          = 4
	MSG_BITFIELD      = 5
	MSG_REQUEST       = 6
	MSG_PIECE         = 7
	MSG_CANCEL        = 8
)

const DefaultBlockSize = 16 * 1024 // 16 KiB

var Err_Unexpected_Message = fmt.Errorf("didn't Recieve expected message")

type PeerMessage struct {
	Length  uint32
	ID      byte
	Payload []byte
}

func NewPeerMessage(id byte, payload []byte) *PeerMessage {
	return &PeerMessage{
		Length:  uint32(len(payload) + 1), // Add 1 for the message ID
		ID:      id,
		Payload: payload,
	}
}

func ReadPeerMessage(reader io.Reader) (*PeerMessage, error) {
	header := make([]byte, 5)

	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[:4])
	id := header[4]

	if length == 0 {
		return &PeerMessage{Length: 0, ID: id, Payload: nil}, nil
	}

	payload := make([]byte, length-1) // Subtract 1 for the message ID
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, err
	}

	return &PeerMessage{Length: length, ID: id, Payload: payload}, nil
}

// WritePeerMessage writes a PeerMessage to a connection.
func WritePeerMessage(conn io.Writer, message *PeerMessage) error {
	// Create a buffer for the message header
	header := make([]byte, 5)
	binary.BigEndian.PutUint32(header[:4], message.Length)
	header[4] = message.ID

	// Write the header and payload to the connection
	_, err := conn.Write(header)
	if err != nil {
		return err
	}
	if len(message.Payload) > 0 {
		_, err := conn.Write(message.Payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func WaitFor(messageType int, conn net.Conn) (*PeerMessage, error) {
	respMsg, err := ReadPeerMessage(conn)
	if err != nil {
		return nil, err
	}
	if respMsg.ID != byte(messageType) {
		fmt.Println(respMsg)
		return nil, Err_Unexpected_Message
	}

	return respMsg, nil
}
