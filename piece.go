package main

//import ()

const BLOCKSIZE int = 16384 //32768

const (
	Undownloaded = iota
	Pending
	Downloaded
)

type Piece struct {
	index      int
	data       []byte
	numBlocks  int
	blocks     []*Block //map[int]*Block
	chanBlocks chan *Block
	peer       *Peer
	hash       [20]byte
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

// parsePieces parses the big wacky string of sha-1 hashes int
// the Info list of
func (info *TorrentInfo) parsePieces() {
	info.cleanPieces()
	// TODO: set this dynamically
	numBlocks := info.PieceLength / int64(BLOCKSIZE)
	info.BlocksPerPiece = int(numBlocks)

	piecesLength := len(info.Pieces)
	info.PieceList = make([]*Piece, 0, piecesLength/20)
	for i := 0; i < piecesLength; i = i + 20 {
		j := i + 20
		piece := Piece{size: info.PieceLength, numBlocks: int(numBlocks)}
		piece.chanBlocks = make(chan *Block)
		//piece.blocks = make(map[int]*Block)
		piece.blocks = make([]*Block, numBlocks)
		// Copy to next 20 into Piece Hash
		copy(piece.hash[:], info.Pieces[i:j])
		piece.length = int(info.PieceLength)
		piece.index = len(info.PieceList)
		//piece.hex = fmt.Sprintf("%x", piece.hash)
		info.PieceList = append(info.PieceList, &piece)
		go piece.checkPieceCompletion()
	}

}
