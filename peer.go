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
		debugger.Println(err)
		return err
	}

	hs := NewHandShake(p.meta, conn)
	pId, err := hs.ShakeHands()
	if err != nil {
		debugger.Println(err)
		return err
	}
	p.Id = pId
	p.Shaken = true
	logger.Println("Connected to Peer: ", pId)

	return nil

}

func (p *Peer) ListenToPeer() {
	logger.Printf("Peer %s : starting Listen\n", p.PeerId)
	// handshake is already authed
	for {
		length := make([]byte, 4)
		n, err := io.ReadFull(p.Conn, length)
		if err != nil {
			debugger.Println("Error Reading, Stopping")
			// TODO: Stop connection
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		n, err = io.ReadFull(p.Conn, payload)
		if err != nil {
			debugger.Println("Error Reading Payload")
			// TODO: Stop connection
		}
		logger.Printf("For %s length I got %d len", length, n)
		logger.Println(string(payload))
	}

}
