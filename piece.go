package main

//import ()

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
	status     int
	//hex        string // Not used
	size   int64
	have   bool
	length int
}

type Block struct {
	piece      *Piece // Not necessary?
	offset     int
	length     int  // Not necessary?
	downloaded bool // Not necessary?
	data       []byte
}

func (b *Block) String() string {
	return (string(b.offset) + " ")
}
