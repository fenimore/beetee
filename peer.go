package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Peer is the basic unit of other.
type Peer struct {
	ip   string
	port uint16
	id   string
	addr string
	// Connection
	conn net.Conn
	// Status Chan
	stopping chan bool
	// Status
	alive      bool
	interested bool
	choked     bool
	choking    bool
	choke      sync.WaitGroup // NOTE: Use this?
	// Messages
	//sendChan chan []byte
	//recvChan chan []byte
	// Piece Data
	bitfield map[int]bool
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
		peer := Peer{
			ip:      ip.String(),
			port:    port,
			addr:    fmt.Sprintf("%s:%d", ip.String(), port),
			choking: true,
			choked:  true,
		}
		Peers = append(Peers, &peer)
	}
}

func (p *Peer) ConnectPeer() error {
	logger.Printf("Connecting to %s", p.addr)
	// Connect to address
	// TODO: Set deadline
	conn, err := net.DialTimeout("tcp", p.addr,
		time.Second*10)
	if err != nil {
		return err
	}
	p.conn = conn
	// NOTE: Does io.Readfull Block?
	err = p.sendHandShake()
	if err != nil {
		return err
	}
	p.alive = true
	logger.Printf("Connected to %s at %s", p.id, p.addr)
	p.choke.Add(1)
	recv := make(chan []byte)
	go p.ListenPeer(recv)
	go p.DecodeMessages(recv)
	return nil
}

// ListenPeer reads from socket.
func (p *Peer) ListenPeer(recv chan<- []byte) {
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.conn, length)
		if err != nil {
			// EOF
			debugger.Printf("Error %s with %s", err, p.id)
			p.stopping <- true
			return
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.conn, payload)
		if err != nil {
			debugger.Printf("Error %s with %s", err, p.id)
			p.stopping <- true
			return
		}
		recv <- payload
		//p.recvChan <- payload
	}
}
