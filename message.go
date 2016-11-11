package main

import "net"
import "bufio"
import "errors"
import "io"
import "bytes"

// 19 bytes

const (
	Choke = iota
	Unchoke
	Interested
	NotInterested
	Have
	BitField
	Request
	Piece
	Cancel
	Port
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
func NewHandShake(meta TorrentMeta, c net.Conn) *HandShake {
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
