package main

const BLOCKSIZE int = 16384

type Piece struct {
	index     int // redundant
	data      []byte
	numBlocks int
	blocks    []*Block
	peer      *Peer
	hash      [20]byte
	size      int64
	have      bool
}

type Block struct {
	piece      *Piece
	offset     int
	length     int
	downloaded bool
	data       []byte
}

// parsePieces parses the big wacky string of sha-1 hashes int
// the Info list of
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	numBlocks := info.PieceLength / int64(BLOCKSIZE)
	len := len(info.Pieces)
	info.PieceList = make([]*Piece, 0, len/20)
	for i := 0; i < len; i = i + 20 {
		j := i + 20
		piece := Piece{size: info.PieceLength, numBlocks: int(numBlocks)}
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		info.PieceList = append(info.PieceList, &piece)
	}
}

func (p *Piece) countBlocks() int {
	//p.size
	return 1
}
