package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
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
	Alive      bool
	Interested bool
	Choked     bool
	Stop       chan bool // TODO: What?
	ChokeWg    sync.WaitGroup
	ListenWg   sync.WaitGroup
	// What the Peer Has, index wise
	Bitfield map[*Piece]bool
	has      map[uint32]bool
}

// parsePeers is a http response gotten from
// the tracker; parse the peers byte message
// and put to global Peers slice.
func (r *TrackerResponse) parsePeers() {
	var start int
	for idx, val := range r.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}
	p := r.Peers[start:]
	// A peer is represented in six bytes
	// four for ip and two for port
	for i := 0; i < len(p); i = i + 6 {
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{Ip: ip.String(), Port: port}
		peer.Choked = true
		// TODO: Set bitfield values to none
		Peers = append(Peers, &peer)
	}
}

// ConnectToPeer establishs handshake (as the shaker,
// not the shakee).
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
	p.Alive = true
	p.ChokeWg.Add(1)
	logger.Println("Connected to Peer: ", p.Id)
	// TODO: Keep alive loop in goroutine
	return nil

}

// ListenToPeer reads from socket connection
// messages to the decoder.
// Important: Connect first
func (p *Peer) ListenToPeer() {
	logger.Printf("Peer %s : starting to Listen\n", p.Id)
	// Listen Loop
	p.ListenWg.Done() // No I'm listening, go ahead and send the status message.
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.Conn, length)
		//debugger.Println(length)
		if err != nil {
			debugger.Printf("Error Reading Length %s, Stopping: %s", p.Id, err)
			p.Alive = false
			p.Conn.Close()
			return
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.Conn, payload)
		if err != nil {
			debugger.Printf("Error Reading Payload %s, Stopping: %s", p.Id, err)
			// TODO: Stop connection
			//p.Stop <- true
			p.Alive = false
			p.Conn.Close()
			return
		}
		go p.decodeMessage(payload)
	}

}
