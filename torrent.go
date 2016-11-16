// beetee Parse package.
package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
	"os"
)

// TorrentMeta is the overarching torrent struct
// and it's info subordinate houses the piece list
type TorrentMeta struct {
	Announce     string        `bencode:"announce"`
	AnnounceList [][]string    `bencode:"announce-list"`
	CreatedBy    string        `bencode:"created by"`
	CreationDate int64         `bencode:"creation date"`
	Comment      string        `bencode:"comment"`
	Encoding     string        `bencode:"encoding"`
	UrlList      []string      `bencode:"url-list"`
	InfoBytes    bencode.Bytes `bencode:"info"`
	Info         *TorrentInfo
	InfoHash     [20]byte
	InfoHashEnc  string
}

func (t *TorrentMeta) String() string {
	return t.Announce
}

// TorrentInfo for single file torrent
// TODO: add support for multiple files
type TorrentInfo struct {
	Length      int64         `bencode:"length"`
	Name        string        `bencode:"name"`
	PieceLength int64         `bencode:"piece length"`
	Pieces      bencode.Bytes `bencode:"pieces"` // The concatnated Bytes
	Private     int64         `bencode:"private"`
	//PieceList      []*Piece
	BlocksPerPiece int
	// md5sum for single files
	// files for multiple files
	// path for multiple files
	//Files       string `bencode:"file"`
	// Concatenation of all 20 byte SHA1
}

// ParseTorrent parses a torrent file.
func ParseTorrent(file string) (*TorrentMeta, error) {
	var data TorrentMeta
	f, err := os.Open(file)
	if err != nil {
		return &data, err
	}
	defer f.Close()
	// Parse the File
	dec := bencode.NewDecoder(f)
	err = dec.Decode(&data)

	if err != nil {
		return &data, err
	}

	// Parse the Info Dictionary
	reader := bytes.NewReader(data.InfoBytes)
	dec = bencode.NewDecoder(reader)
	dec.Decode(&data.Info)
	// Compute the info_hash
	data.InfoHash = sha1.Sum(data.InfoBytes)
	data.InfoHashEnc = UrlEncode(data.InfoHash)
	data.Info.parsePieces()
	return &data, nil
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
		debugger.Println(info.Pieces[i:j])
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

// cleanPieces because the encoder includes the lenght of the
// string of bytes
func (info *TorrentInfo) cleanPieces() {
	var idx int
	for i, val := range info.Pieces {
		if val == ':' {
			idx = i + 1
			break
		}
	}
	info.Pieces = info.Pieces[idx:]
}

func UrlEncode(hash [20]byte) string {
	var enc string
	for _, b := range hash {
		switch {
		case b >= 97 && b <= 122: // lower case
			enc += string(b)
			continue
		case b >= 65 && b <= 90: // upper case
			enc += string(b)
			continue
		case b >= 48 && b <= 57: // numbers
			enc += string(b)
			continue
		case b == 46 || b == 45:
			enc += string(b)
			continue
		case b == 95:
			enc += string(b)
			continue
		case b == 126:
			enc += string(b)
			continue
		default:
			enc += "%" + fmt.Sprintf("%x", []byte{b})
		}
	}
	return enc
}
