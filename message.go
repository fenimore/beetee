package main

import "net"
import "bufio"
import "errors"
import "io"
import "bytes"

// 19 bytes

const (
	ChokeMsg = iota
	UnchokeMsg
	InterestedMsg
	NotInterestedMsg
	HaveMsg
	BitFieldMsg
	RequestMsg
	BlockMsg // rather than PieceMsg
	CancelMsg
	PortMsg
)

type HandShake struct {
	pstr     []byte
	pstrlen  uint8 // fit into one byte
	reserved []byte
	infoHash []byte
	peerId   []byte
	conn     net.Conn
	writer   *bufio.Writer
}

///<pstrlen><pstr><reserved><info_hash><peer_id>
// 68 bytes long.
func NewHandShake(meta *TorrentMeta, c net.Conn) *HandShake {
	handshake := new(HandShake)
	handshake.pstr = []byte{'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l'}
	handshake.pstrlen = (uint8)(len(handshake.pstr))
	handshake.reserved = make([]byte, 8) // 8 empty bytes

	handshake.infoHash = meta.InfoHash[:]
	handshake.peerId = peerId[:]

	handshake.conn = c

	handshake.writer = bufio.NewWriter(handshake.conn)

	return handshake
}

func (h *HandShake) ShakeHands() (string, error) {
	var n int
	var err error
	err = h.writer.WriteByte(h.pstrlen)
	if err != nil {
		return "", err
	}
	n, err = h.writer.Write(h.pstr)
	if err != nil || n != len(h.pstr) {
		return "", err
	}
	n, err = h.writer.Write(h.reserved)
	if err != nil || n != len(h.reserved) {
		return "", err
	}
	n, err = h.writer.Write(h.infoHash)
	if err != nil || n != len(h.infoHash) {
		return "", err
	}
	n, err = h.writer.Write(h.peerId)
	if err != nil || n != len(h.peerId) {
		return "", err
	}
	err = h.writer.Flush()
	if err != nil {
		return "", err
	}

	// The response handshake
	shake := make([]byte, 68)
	n, err = io.ReadFull(h.conn, shake)
	if err != nil {
		return "", err
	}
	// TODO: Check for Length
	if !bytes.Equal(shake[1:20], h.pstr) {
		return "", errors.New("Protocol does not match")
	}
	if !bytes.Equal(shake[28:48], h.infoHash) {
		return "", errors.New("InfoHash Does not match")
	}

	return string(shake[48:68]), nil
}

func (p *Peer) decodeMessage(payload []byte) {
	// first byte is msg type
	if len(payload) < 1 {
		return
	}
	msg := payload[1:]
	switch payload[0] {
	case ChokeMsg:
		logger.Println("Choked", msg)
	case UnchokeMsg:
		logger.Println("UnChoke", msg)
	case InterestedMsg:
		logger.Println("Interested", msg)
	case NotInterestedMsg:
		logger.Println("NotInterested", msg)
	case HaveMsg:
		logger.Println("Have", msg)
	case BitFieldMsg:
		logger.Println("Bitfield", msg)
	case RequestMsg:
		logger.Println("Request", msg)
	case BlockMsg:
		logger.Println("Piece", msg)
	case CancelMsg:
		logger.Println("Payload", msg)
	case PortMsg:
		logger.Println("Port", msg)
	}

}

func (p *Peer) sendStatusMessage(msg int) error {
	logger.Println("Sending Status Message: ", msg)
	var err error
	writer := bufio.NewWriter(p.Conn)
	if msg == -1 { // keep alive, do nothing TODO: add ot iota
		_, err = writer.Write([]byte{'0', '0', '0', '0'})

	} else {
		_, err = writer.Write([]byte{(uint8)(0),
			(uint8)(0), (uint8)(0), (uint8)(1)})
	}
	if err != nil {
		return err
	}

	// Format
	//<len=0001><id=0>
	switch msg {
	case ChokeMsg:
		err = writer.WriteByte((uint8)(0))
	case UnchokeMsg:
		err = writer.WriteByte((uint8)(1))
	case InterestedMsg:
		err = writer.WriteByte((uint8)(2))
	case NotInterestedMsg:
		err = writer.WriteByte((uint8)(3))
	}
	if err != nil {
		return err
	}
	//debugger.Println("Sending", writer)
	writer.Flush()
	return nil
}
