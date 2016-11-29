package main

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	// Status
	alive      bool
	interested bool
	choked     bool
	chokeWg    sync.WaitGroup
	// Piece Data
	bitfield []byte
}

func (p *Peer) Connect() (*net.Conn, error) {
	logger.Printf("Connecting to %s", p.addr)

	// Connect to address
	conn, err := net.DialTimeout("tcp", p.addr, time.Second*10)
	if err != nil {
		return conn, err
	}
	logger.Printf("Connected to %s at %s", p.id, p.addr)
	return conn, err
}

func (p *Peer) HandShake(conn *net.Conn, info *TorrentMeta) error {
	// The response handshake
	shake := make([]byte, 68)
	hs := HandShake(info)
	conn.Write(hs[:])

	_, err := io.ReadFull(p.conn, shake)
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

// readMessage reads from connection, It blocks
func ReadMessage(conn net.Conn) ([]byte, error) {
	var err error
	// NOTE: length is 4 byte big endian
	length := make([]byte, 4)
	_, err = io.ReadFull(conn, length)
	if err != nil {
		return nil, err
	}

	if binary.BigEndian.Uint32(length) < 1 {
		return nil, nil // Keep Alive
	}

	payload := make([]byte, binary.BigEndian.Uint32(length))
	_, err = io.ReadFull(conn, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func handleMessage(payload []byte) error {
	if len(payload) < 1 {
		return nil // NOTE, Keep alive was recv
	}

	switch payload[0] {
	case ChokeMsg:
		// TODO: Set Peer unchoke
		p.choked = false
		p.chokeWg.Done()
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
		idx := DecodeHaveMessage(payload)
		logger.Printf("Recv: %s sends have %v for Piece %d",
			p.id, payload[1:], idx)
	case BitFieldMsg:
		// TODO: Decode into slice?
		logger.Printf("Recv: %s sends bitfield", p.id)
	case RequestMsg:
		logger.Printf("Recv: %s sends request %s", p.id, payload)
	case BlockMsg: // Officially "Piece" message
		// TODO: Remove this message, as they are toomuch
		logger.Printf("Recv: %s sends block", p.id)
		b := DecodePieceMessage(payload)
		Pieces[b.index].chanBlocks <- b
		if len(Pieces[b.index].chanBlocks) == cap(Pieces[b.index].chanBlocks) {
			Pieces[b.index].VerifyPiece() // FIXME: Goroutine?
		}
	case CancelMsg:
		logger.Printf("Recv: %s sends cancel %s", p.id, payload)
	case PortMsg:
		logger.Printf("Recv: %s sends port %s", p.id, payload)
	default:
		break

	}
	return nil
}

func (p *Peer) openChannel(in chan []byte) (chan []byte, chan []byte) {
	out := make(chan []byte)
	halt := make(chan []byte)
	go func() {
		err, conn := p.Connect()
		if err != nil {
			close(out)
		}
		for {
			select {
			case msg := <-in:
				conn.Write(msg)
			case data := <-conn:
				//x
			case end := <-halt:
				close(out)
			}
		}
	}()
	return out, halt
}
