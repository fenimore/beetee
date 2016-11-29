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
	info *TorrentMeta
	conn net.Conn
	// State
	sync.Mutex
	choke      bool
	interested bool
	alive      bool
	bitfield   []byte
	// Chan
	in   chan []byte
	out  chan []byte
	halt chan struct{}
}

func (p *Peer) Connect() error {
	logger.Printf("Connecting to %s", p.addr)

	// Connect to address
	conn, err := net.DialTimeout("tcp", p.addr, time.Second*10)
	if err != nil {
		return err
	}
	p.conn = conn
	logger.Printf("Connected to %s at %s", p.id, p.addr)
	return err
}

func (p *Peer) HandShake() error {
	// The response handshake
	shake := make([]byte, 68)
	hs := HandShake(p.info)
	p.conn.Write(hs[:])

	_, err := io.ReadFull(p.conn, shake)
	if err != nil {
		return err
	}

	// TODO: Check for Length
	if !bytes.Equal(shake[1:20], pstr) {
		return errors.New("Protocol does not match")
	}
	if !bytes.Equal(shake[28:48], p.info.InfoHash[:]) {
		return errors.New("InfoHash Does not match")
	}

	p.id = string(shake[48:68])
	return nil
}

// readMessage reads from connection, It blocks
func (p *Peer) readMessage() ([]byte, error) {
	var err error
	// NOTE: length is 4 byte big endian
	length := make([]byte, 4)
	_, err = io.ReadFull(p.conn, length)
	if err != nil {
		return nil, err
	}

	if binary.BigEndian.Uint32(length) < 1 {
		return nil, nil // Keep Alive
	}

	payload := make([]byte, binary.BigEndian.Uint32(length))
	_, err = io.ReadFull(p.conn, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (p *Peer) handleMessage(payload []byte, waiting, choked, ready chan<- *Peer) error {
	if len(payload) < 1 {
		return nil // NOTE, Keep alive was recv
	}

	switch payload[0] {
	case ChokeMsg:
		// TODO: Set Peer unchoke
		// Halt?
		//close(p.halt)
		p.Lock()
		p.choke = true
		p.Unlock()
		choked <- p
		logger.Printf("Recv: %s sends choke", p.id)
	case UnchokeMsg:
		p.Lock()
		p.choke = false
		p.Unlock()
		ready <- p
		logger.Printf("Recv: %s sends unchoke", p.id)
	case InterestedMsg:
		p.interested = true
		logger.Printf("Recv: %s sends interested", p.id)
	case NotInterestedMsg:
		p.interested = false
		logger.Printf("Recv: %s sends uninterested", p.id)
	case HaveMsg:
		idx := DecodeHaveMessage(payload)
		// TODO: Update bitfield
		logger.Printf("Recv: %s sends have %v for Piece %d",
			p.id, payload[1:], idx)
	case BitFieldMsg:
		// TODO: Decode into slice?
		logger.Printf("Recv: %s sends bitfield", p.id)
	case RequestMsg:
		logger.Printf("Recv: %s sends request %s", p.id, payload)
	case BlockMsg: // Officially "Piece" message
		// TODO: Remove this message, as they are toomuch
		//logger.Printf("Recv: %s sends block", p.id)
		b := DecodePieceMessage(payload)
		d.Pieces[b.index].chanBlocks <- b
		if len(d.Pieces[b.index].chanBlocks) == cap(d.Pieces[b.index].chanBlocks) {
			d.Pieces[b.index].VerifyPiece() // FIXME: Goroutine?
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

func (p *Peer) spawnPeerReader(disconnected chan<- *Peer) chan struct{} {
	halt := make(chan struct{})
	go func() {
		for {
			select {
			case <-halt:
				disconnected <- p
				p.conn.Close()
				debugger.Println("Halt closes Peer", p.id)
				break
			default:
				msg, err := p.readMessage()
				if err != nil {
					logger.Println(err)
					close(halt)
					break
				}
				p.out <- msg
			}
		}
	}()
	return halt
}

func (p *Peer) spawnPeerHandler(waiting, choked, ready, disconnected chan<- *Peer) {
	p.in = make(chan []byte)
	p.out = make(chan []byte)
	p.halt = make(chan struct{})
	go func() {
		err := p.Connect()
		if err != nil {
			close(p.out)
			debugger.Println("Connection Fails", err)
			return
		}

		err = p.HandShake()
		if err != nil {
			close(p.out)
		}

		closeReader := p.spawnPeerReader(disconnected)
		choked <- p
		for {
			select {
			case msg := <-p.in:
				p.conn.Write(msg)
			case msg := <-p.out:
				p.handleMessage(msg, waiting, choked, ready)
			case <-p.halt:
				close(closeReader)
				close(p.out)
				break
			}
		}
	}()
}

func (p *Peer) spawnPieceRequest(piece int, info *TorrentInfo) chan *Piece {
	out := make(chan *Piece)
	go func() {
		blocksPerPiece := int(info.PieceLength) / BLOCKSIZE
		for offset := 0; offset < blocksPerPiece; offset++ {
			RequestMessage(uint32(piece), offset*BLOCKSIZE)

		}

	}()
	return out
}
