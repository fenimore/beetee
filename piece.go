package main

import (
	"sync"
)

const (
	blocksize int = 16384
)

// Piece is the general unit that files are divided into.
type Piece struct {
	index      int
	data       []byte
	size       int64
	hash       [20]byte
	chanBlocks chan *Block
	verified   bool
	//hex string // NOTE: no need
	pending sync.WaitGroup
	success chan bool // For when all blocks are there
}

// Block struct will always have constant size, 16KB.
type Block struct {
	index  uint32 // NOTE: piece index
	offset uint32
	data   []byte
}

// parsePieces constructs the Piece list from
// the torrent file.
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	numberOfBlocks := info.PieceLength / int64(blocksize)
	// NOTE: Pieces are global variable of all pieces
	Pieces = make([]*Piece, 0, len(info.Pieces)/20)
	for i := 0; i < len(info.Pieces); i = i + 20 {
		j := i + 20 // NOTE: j is hash end
		piece := Piece{
			size:       info.PieceLength,
			chanBlocks: make(chan *Block, numberOfBlocks), // numberOfBlocks
			data:       make([]byte, info.PieceLength),
			index:      len(Pieces),
			verified:   false,
			success:    make(chan bool),
			//hash:       fmt.Sprintf("%x", piece.hash),
		}
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		Pieces = append(Pieces, &piece)
	}
}
