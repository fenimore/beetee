package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Peer struct {
	meta *TorrentMeta // Whence it came

	PeerId string `bencode:"peer id"` // Bencoding not being used
	Ip     string `bencode:"ip"`
	Port   uint16 `bencode:"port"`
	// After connection
	Shaken bool
	Conn   net.Conn
	Id     string
	// Peer Values
	Interesed bool
	Choked    bool
	Bitfield  []bool
}

func (p *Peer) ConnectToPeer() error {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.Port)
	logger.Println("Connecting to Peer: ", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	hs := NewHandShake(p.meta, conn)
	pId, err := hs.ShakeHands()
	if err != nil {
		return err
	}
	p.Id = pId
	p.Shaken = true
	p.Conn = conn
	logger.Println("Connected to Peer: ", pId)

	return nil

}

func (p *Peer) ListenToPeer() {
	logger.Printf("Peer %s : starting Listen\n", string(p.PeerId))
	// handshake is already authed
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.Conn, length)
		if err != nil {
			debugger.Println("Error Reading, Stopping")
			// TODO: Stop connection
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.Conn, payload)
		if err != nil {
			debugger.Println("Error Reading Payload")
			// TODO: Stop connection
		}
		go p.decodeMessage(payload)
	}

}

func (p *Peer) decodeMessage(payload []byte) {
	// first byte is msg type
	msg := payload[1:]
	switch payload[0] {
	case Choke:
		logger.Println("Choked", msg)
	case Unchoke:
		logger.Println("UnChocke", msg)
	case Interested:
		logger.Println("Interested", msg)
	case NotInterested:
		logger.Println("NotInterested", msg)
	case Have:
		logger.Println("Have", msg)
	case BitField:
		logger.Println("Bitfield", msg)
	case Request:
		logger.Println("Request", msg)
	case Piece:
		logger.Println("Piece", msg)
	case Cancel:
		logger.Println("Payload", msg)
	case Port:
		logger.Println("Port", msg)
	}

}
