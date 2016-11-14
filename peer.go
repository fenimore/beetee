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
	// Buffer
	// TODO: make writer?
	// After connection
	Shaken bool
	Conn   net.Conn
	Id     string
	// Peer Status
	Interesed bool
	Choked    bool
	Bitfield  []bool
	// What the Peer Has, index wise
	has map[uint32]bool
}

func (p *Peer) ConnectToPeer() error {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.Port)
	logger.Println("Connecting to Peer: ", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	p.Conn = conn
	p.has = make(map[uint32]bool)
	err = p.ShakeHands()
	if err != nil {
		return err
	}
	logger.Println("Connected to Peer: ", p.Id)

	return nil

}

//<len>   <id><payload>
func (p *Peer) ListenToPeer() {
	// Handshake
	err := p.ConnectToPeer()
	if err != nil {
		debugger.Println("Error Connecting to  %s: %s", p.Id, err)
		return
	}
	logger.Printf("Peer %s : starting to Listen\n", p.Id)
	// handshake is already authed
	for {
		length := make([]byte, 4)
		_, err = io.ReadFull(p.Conn, length)
		//debugger.Println(length)
		if err != nil {
			debugger.Println("Error Reading, Stopping", err)
			p.sendStatusMessage(-1)
			// TODO: Stop connection
			// This is typicaly io.EOF error
			return
		} //make([]array,1)
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.Conn, payload)
		if err != nil {
			debugger.Println("Error Reading Payload", err)
			// TODO: Stop connection
			return
		}
		go p.decodeMessage(payload)
	}
}
