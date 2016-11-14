package main

type Piece struct {
	index  int // redundant
	data   []byte
	blocks []*Block
	peer   *Peer
	hash   [20]byte
	size   int64
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
	len := len(info.Pieces) / 20
	info.PieceList = make([]*Piece, len)
	for i := 0; i < len; i = i + 20 {
		j := i + 20
		piece := Piece{}
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		piece.size = info.PieceLength
		info.PieceList = append(info.PieceList, &piece)
	}
}
