package main

import (
	"net"
)

const (
	blocksize int = 16384
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
	stopping   chan struct{}
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
	confirmed  bool
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
			ip:   ip.String(),
			port: port,
			addr: fmt.Sprintf("%s:%d", p.Ip, p.Port),
			choking: true
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
		log.Println("The block channel is not full")
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
	p.confirmed = true
	log.Printf("Piece at %d is successfully written", p.index)
	ioChan <- p
}

func OutputData() {
	data := make([]byte, Torrent.Info.Length)
	for {
		piece := <-ioChan
	}
}
