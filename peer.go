package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Peer is the basic unit of other.
type Peer struct {
	ip   string
	port uint16
	id   string
	addr string
	// Status
	alive      bool
	interested bool
	choked     bool
	// Piece Data
	bitfield []byte
}

// parsePeers is a http response gotten from
// the tracker; parse the peers byte message
// and put to global Peers slice.
func ParsePeers(r TrackerResponse) []*Peer {
	var start int
	var peers []*Peer
	for idx, val := range r.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}
	p := r.Peers[start:]
	// A peer is represented in six bytes
	// four for ip and two for port
	bitCap := len(Pieces) / 8
	if len(Pieces)%8 != 0 {
		bitCap += 1
	}

	for i := 0; i < len(p); i = i + 6 {
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{
			ip:       ip.String(),
			port:     port,
			addr:     fmt.Sprintf("%s:%d", ip.String(), port),
			choked:   true,
			bitfield: make([]byte, bitCap),
		}
		peers = append(peers, &peer)
	}
	return peers
}

func (p *Peer) Connect() (net.Conn, error) {
	logger.Printf("Connecting to %s", p.addr)

	// Connect to address
	conn, err := net.DialTimeout("tcp", p.addr, time.Second*10)
	if err != nil {
		return conn, err
	}

	logger.Printf("Connected to %s at %s", p.id, p.addr)
	return conn, err
}

func (p *Peer) HandShake(conn net.Conn, info *TorrentMeta) error {
	// The response handshake
	shake := make([]byte, 68)
	hs := HandShake(info)
	conn.Write(hs[:])

	_, err := io.ReadFull(conn, shake)
	if err != nil {
		return err
	}

	// TODO: Check for Length
	if !bytes.Equal(shake[1:20], pstr) {
		return errors.New("Protocol does not match")
	}
	if !bytes.Equal(shake[28:48], info.InfoHash[:]) {
		return errors.New("InfoHash Does not match")
	}

	p.id = string(shake[48:68])
	return nil
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
		logger.Printf("Recv: %s sends choke", p.id)
	case UnchokeMsg:
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
