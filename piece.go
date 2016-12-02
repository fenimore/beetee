package main

import (
	"crypto/sha1"
	"sync"
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
	size   int // typicall BLOCKSIZE, except for last
}

// parsePieces constructs the Piece list from
// the torrent file.
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	numberOfBlocks := info.PieceLength / int64(BLOCKSIZE)
	// NOTE: Pieces are global variable of all pieces
	d.Pieces = make([]*Piece, 0, len(info.Pieces)/20)
	for i := 0; i < len(info.Pieces); i = i + 20 {
		j := i + 20 // NOTE: j is hash end
		piece := Piece{
			size:     info.PieceLength,
			index:    len(d.Pieces),
			verified: false,
			success:  make(chan bool),
		}
		// Last piece has different amount of blocks
		if i+20 >= len(info.Pieces) {
			piece.chanBlocks = make(chan *Block, info.lastPieceBlockCount())
			piece.size = info.lastPieceSize()
			piece.data = make([]byte, piece.size)
			//logger.Println(piece.size, cap(piece.chanBlocks), BLOCKSIZE)
		} else {
			piece.chanBlocks = make(chan *Block, numberOfBlocks)
			piece.data = make([]byte, info.PieceLength)
		}

		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		d.Pieces = append(d.Pieces, &piece)
	}
	bitCap := len(d.Pieces) / 8
	if len(d.Pieces)%8 != 0 {
		bitCap += 1
	}

	d.bitfield = make([]byte, bitCap)
}

func (info *TorrentInfo) lastPieceBlockCount() int64 {
	if info.Length%info.PieceLength == 0 {
		return info.PieceLength / int64(BLOCKSIZE)
	}
	pieceCount := (info.Length % info.PieceLength) / int64(BLOCKSIZE)
	if pieceCount == 0 {
		return 1
	}
	return pieceCount
}

func (info *TorrentInfo) lastBlockSize() int64 {
	if info.Length%info.PieceLength == 0 {
		return (info.PieceLength / int64(BLOCKSIZE)) * BLOCKSIZE
	}

	return 0

}

func (info *TorrentInfo) lastPieceSize() int64 {
	if info.Length%info.PieceLength != 0 {
		return info.Length % info.PieceLength
	}
	return info.PieceLength

}

func (p *Piece) VerifyPiece() {
	for {
		b := <-p.chanBlocks
		copy(p.data[int(b.offset):int(b.offset)+b.size],
			b.data)
		if len(p.chanBlocks) < 1 {
			break
		}
	}
	if p.hash != sha1.Sum(p.data) {
		debugger.Printf(
			"Error with piece of size %d,\n"+
				"the hash is %x, and what I got is %x",
			p.size, p.hash, sha1.Sum(p.data))
		p.data = nil
		p.data = make([]byte, p.size)
		logger.Printf("Unable to Write Blocks to Piece %d", p.index)
		return
	}
	p.verified = true
	//logger.Printf("Piece at %d is Collected", p.index)
	// TODO: Update personal bitfield
	// TODO: Send have to peers
	p.success <- true // FIXME: Keep?
}

func UpdateBitfield() []byte {
	// result := make([]bool, len(d.Pieces)/8)
	// bitfield := msg[1:]
	// // For each byte, look at the bits
	// // NOTE: that is 8 * 8
	// for i := 0; i < len(bitfield); i++ {
	//	for j := 0; j < 8; j++ {
	//		index := i*8 + j
	//		if index >= len(d.Pieces) {
	//			break // Hit padding bits
	//		}
	//		byte := bitfield[i]              // Within bytes
	//		bit := (byte >> uint32(7-j)) & 1 // some shifting
	//		result[index] = bit == 1         // if bit is true
	//	}
	// }
	// return result
	return nil
}
