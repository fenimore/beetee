package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"
)

const (
	blocksize int = 16384
)

const (
	ChokeMsg = iota
	UnchokeMsg
	InterestedMsg
	NotInterestedMsg
	HaveMsg
	BitFieldMsg
	RequestMsg
	BlockMsg // rather than PieceMsg
	CancelMsg
	PortMsg
)

var (
	pieceChan chan *Piece
	peerChan  chan *Peer
	ioChan    chan *Piece
)

// PeerInfo is parsed form the tracker
type PeerInfo struct {
	ip   string `bencode:"ip"`
	port uint16 `bencode:"port"`
}

// Peer is the basic unit of other.
type Peer struct {
	ip   string
	port uint16
	id   string
	addr string
	// Connection
	conn net.Conn
	// Status
	//	stopping   chan struct{}
	choking    chan struct{}
	alive      bool
	interested bool
	choked     bool
	choking    bool
	// Messages
	sendChan chan []byte
	recvChan chan []byte
	// Piece Data
	bitfield map[int]bool
}

// Piece is the general unit that files are divided into.
type Piece struct {
	index      int
	data       []byte
	size       int64
	hash       [20]byte
	chanBlocks chan *Block
	verified   bool
	//hex string // NOTE: no need
}

// Block struct will always have constant size, 16KB.
type Block struct {
	index  int32 // NOTE: piece index
	offset int32
	data   []byte
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
			addr:    fmt.Sprintf("%s:%d", p.Ip, p.Port),
			choking: true,
			choked:  true,
		}
		Peers = append(Peers, &peer)
	}
}

// parsePieces constructs the Piece list from
// the torrent file.
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	numberOfBlocks := info.PieceLength / int64(BLOCKSIZE)
	info.BlocksPerPiece = int(numBlocks)
	// NOTE: Pieces are global variable of all pieces
	Pieces = make([]*Piece, 0, len(info.Pieces)/20)
	for i := 0; i < piecesLength; i = i + 20 {
		j := i + 20 // NOTE: j is hash end
		piece := Piece{
			size:       info.PieceLength,
			chanBlocks: make(chan *Block, numberOfBlocks),
			data:       make([]byte, info.PieceLength),
			index:      len(Pieces),
			confirmed:  false,
			//hash:       fmt.Sprintf("%x", piece.hash),
		}
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		Pieces = append(Pieces, &piece)
	}
}

func (p *Peer) decodePieceMessage(msg []byte) {
	if len(msg[8:]) < 1 {
		return
	}
	index := binary.BigEndian.Uint32(msg[:4])
	begin := binary.BigEndian.Uint32(msg[4:8])
	data := msg[8:]
	// Blocks...
	block := &Block{index: index, offset: begin, data: data}
	Pieces[index].chanBlocks <- block

	if len(Pieces[index].chanBlocks) == cap(Pieces[index].chanBlocks) {
		Pieces[index].writeBlocks()
	}
}

func (p *Piece) writeBlocks() {
	if len(p.chanBlocks) < cap(p.chanBlocks) {
		logger.Println("The block channel is not full")
		return
	}
	for {
		b := <-p.chanBlocks // NOTE: b for block
		// Copy block data to p
		// NOTE: If this doesn't work,
		// Go back to old manner of using a indexed in
		// block array
		copy(p.data[b.offset:b.offset+blocksize],
			b.data)
		if len(p.chanBlocks) < 1 {
			break
		}
	}
	if p.hash != sha1.Sum(p.data) {
		p.data = nil
		p.data = make([]byte, p.size)
		log.Println("Unable to Write Blocks to Piece")
		return
	}
	p.verified = true
	logger.Printf("Piece at %d is successfully written", p.index)
	ioChan <- p
}

func (p *Peer) ConnectPeer() error {
	log.Printf("Connecting to %s", p.addr)
	// Connect to address
	conn, err := net.Dial("tcp", p.addr)
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
	return nil
}

// ListenPeer reads from socket.
func (p *Peer) ListenPeer() {
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.conn, length)
		if err != nil {
			// EOF
			debugger.Printf("Error %s with %s", err, p.id)
			p.alive = false
			p.conn.Close()
			return
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.conn, payload)
		if err != nil {
			debugger.Printf("Error %s with %s", err, p.id)
			p.alive = false
			p.conn.Close()
			return
		}
		p.recvChan <- payload
	}
}

func (p *Peer) DecodeMessage() {
	for {
		// FIXME: Non blocking
		// FIXME  and send a stopping channel
		// FIXME: and then return
		payload := <-p.recvChan
		msg := payload[1:]
		switch payload[0] {
		case ChokeMsg:
			p.choking = true
			// TODO: waitgroup
			logger.Printf("Recv: %s sends choke", p.id)
		case UnchokeMsg:
			p.chocking = false
			logger.Printf("Recv: %s sends unchoke", p.id)
		case InterestedMsg:
			p.interested = true
			logger.Printf("Recv: %s sends interested", p.id)
		case NotInterestedMsg:
			p.interested = false
			logger.Printf("Recv: %s sends uninterested", p.id)
		case HaveMsg:
			// TODO: Set bitfield
			logger.Printf("Recv: %s sends have %s", p.id, msg)
		case BitFieldMsg:
			//TODO: Parse Bitfield
			logger.Printf("Recv: %s sends bitfield %s",
				p.id, msg)
		case RequestMsg:
			logger.Printf("Recv: %s sends request %s", p.id, msg)
		case BlockMsg: // Officially "Piece" message
			logger.Printf("Recv: %s sends piece", p.id)
			p.decodePieceMessage(msg)
		case CancelMsg:
			logger.Printf("Recv: %s sends cancel %s", p.id, msg)
		case PortMsg:
			logger.Printf("Recv: %s sends port %s", p.id, msg)
		default:
			continue

		}
	}
}

// sendHandShake asks another client to accept your connection.
func (p *Peer) sendHandShake() error {
	///<pstrlen><pstr><reserved><info_hash><peer_id>
	// 68 bytes long.
	var n int
	var err error
	writer := bufio.NewWriter(p.conn)
	// Handshake message:
	pstrlen := byte(19) // or len(pstr)
	pstr := []byte{'B', 'i', 't', 'T', 'o', 'r',
		'r', 'e', 'n', 't', ' ', 'p', 'r',
		'o', 't', 'o', 'c', 'o', 'l'}
	reserved := make([]byte, 8)
	info := Torrent.InfoHash[:]
	id := peerId[:] // my peerId
	// Send handshake message
	err = writer.WriteByte(pstrlen)
	if err != nil {
		return err
	}
	n, err = writer.Write(pstr)
	if err != nil || n != len(pstr) {
		return err
	}
	n, err = writer.Write(reserved)
	if err != nil || n != len(reserved) {
		return err
	}
	n, err = writer.Write(info)
	if err != nil || n != len(info) {
		return err
	}
	n, err = writer.Write(id)
	if err != nil || n != len(id) {
		return err
	}
	err = writer.Flush() // TODO: Do I need to Flush?
	if err != nil {
		return err
	}
	// receive confirmation

	// The response handshake
	shake := make([]byte, 68)
	// TODO: Does this block?
	n, err = io.ReadFull(p.conn, shake)
	if err != nil {
		return err
	}
	// TODO: Check for Length
	if !bytes.Equal(shake[1:20], pstr) {
		return errors.New("Protocol does not match")
	}
	if !bytes.Equal(shake[28:48], info) {
		return errors.New("InfoHash Does not match")
	}
	p.id = string(shake[48:68])
	return nil
}
