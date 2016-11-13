package main

type Piece struct {
	index  int
	data   []byte
	blocks []*Block
	peer   *Peer
	hash   [20]byte
}

type Block struct {
	piece      *Piece
	offset     int
	length     int
	downloaded bool
	data       []byte
}

func ParsePieces(pieces []byte) []*Piece {

	return nil
}
