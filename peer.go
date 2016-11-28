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
	// Status
	sync.Mutex // NOTE: Should be RWMutex?
	alive      bool
	interested bool
	choked     bool

	choke sync.WaitGroup // NOTE: Use this?
	// Messages
	//sendChan chan []byte
	//recvChan chan []byte
	// Piece Data
	bitfield []bool
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
			ip:       ip.String(),
			port:     port,
			addr:     fmt.Sprintf("%s:%d", ip.String(), port),
			choked:   true,
			bitfield: make([]bool, len(Pieces)),
		}
		Peers = append(Peers, &peer)
	}
}

func (p *Peer) ConnectPeer() error {
	logger.Printf("Connecting to %s", p.addr)
	// Connect to address
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
	logger.Printf("Connected to %s at %s", p.id, p.addr)
	p.Lock()
	p.alive = true
	p.choke.Add(1)
	p.Unlock()
	//recv := make(chan []byte)
	go p.ListenPeer() //recv)
	//go p.DecodeMessages(recv)
	return nil
}

// ListenPeer reads from socket.
func (p *Peer) ListenPeer() {
	for {
		select {
		case <-p.stopping:
			debugger.Printf("Peer %s is closing", p.id)
			p.Lock()
			p.conn.Close()
			p.alive = false
			p.Unlock()
			return
		default:
			// do nothing
		}
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
		//recv <- payload
		if len(payload) < 1 {
			continue
		}
		go p.DecodeMessages(payload)
	}
}

func (p *Peer) DecodeMessages(payload []byte) {
	//var payload []byte
	//var msg []byte
	if len(payload) < 1 {
		return
	}
	msg := payload[1:]
	switch payload[0] {
	case ChokeMsg:
		p.choking <- true
		p.Lock()
		p.choked = true
		p.choke.Add(1)
		p.Unlock()
		logger.Printf("Recv: %s sends choke", p.id)
	case UnchokeMsg:
		if p.choked {
			p.Lock()
			p.choked = false
			p.choke.Done()
			p.Unlock()
		}
		logger.Printf("Recv: %s sends unchoke", p.id)
	case InterestedMsg:
		p.interested = true
		logger.Printf("Recv: %s sends interested", p.id)
	case NotInterestedMsg:
		p.interested = false
		logger.Printf("Recv: %s sends uninterested", p.id)
	case HaveMsg:
		p.decodeHaveMessage(msg)
		logger.Printf("Recv: %s sends have %v", p.id, msg)
	case BitFieldMsg:
		p.decodeBitfieldMessage(msg)
		logger.Printf("Recv: %s sends bitfield", p.id)
	case RequestMsg:
		logger.Printf("Recv: %s sends request %s", p.id, msg)
	case BlockMsg: // Officially "Piece" message
		// TODO: Remove this message, as they are toomuch
		//logger.Printf("Recv: %s sends block", p.id)
		p.decodePieceMessage(msg)
	case CancelMsg:
		logger.Printf("Recv: %s sends cancel %s", p.id, msg)
	case PortMsg:
		logger.Printf("Recv: %s sends port %s", p.id, msg)
	default:
		break

	}

}
