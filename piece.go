package main

import (
	"sync"
)

const BLOCKSIZE int = 16384 //32768

const (
	Empty = iota
	Pending
	Full
)

type Piece struct {
	index      int
	data       []byte
	numBlocks  int
	blocks     []*Block //map[int]*Block
	chanBlocks chan *Block
	peer       *Peer
	hash       [20]byte
	// Mutex
	sync.RWMutex     /// Reading is OK
	status       int // Empty Pending or Full
	//hex        string // Not used
	size   int64
	have   bool
	length int
	// WaitGroup
	Pending sync.WaitGroup

	// TODO: Request Timeout
}

type Block struct {
	piece      *Piece // Not necessary?
	offset     int
	length     int  // Not necessary?
	downloaded bool // Not necessary?
	data       []byte
}

// parsePieces parses the big wacky string of sha-1 hashes int
// the Info list of
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	// TODO: set this dynamically
	numBlocks := info.PieceLength / int64(BLOCKSIZE)
	info.BlocksPerPiece = int(numBlocks)

	piecesLength := len(info.Pieces)
	Pieces = make([]*Piece, 0, piecesLength/20)
	for i := 0; i < piecesLength; i = i + 20 {
		j := i + 20
		//debugger.Println(info.Pieces[i:j])
		piece := Piece{size: info.PieceLength, numBlocks: int(numBlocks)}
		piece.chanBlocks = make(chan *Block)
		//piece.blocks = make(map[int]*Block)
		piece.blocks = make([]*Block, numBlocks)
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		piece.length = int(info.PieceLength)
		piece.index = len(Pieces)
		//piece.hex = fmt.Sprintf("%x", piece.hash)
		Pieces = append(Pieces, &piece)
	}
}

func (b *Block) String() string {
	return (string(b.offset) + " ")
}
