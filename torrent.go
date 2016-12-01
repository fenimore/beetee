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
	PeerId       [20]byte // NOTE: my peer ID
}

func (t *TorrentMeta) String() string {
	return t.Announce
}

// TorrentInfo for single file torrent
// TODO: add support for multiple files
type TorrentInfo struct {
	SingleFile  bool
	PieceLength int64         `bencode:"piece length"`
	Pieces      bencode.Bytes `bencode:"pieces"` // The concatnated Bytes
	Private     int64         `bencode:"private"`
	// Single File
	Length int64  `bencode:"length"`
	Name   string `bencode:"name"`
	// Multiple Files
	FilesBytes bencode.Bytes `bencode:"files"`
	Files      []TorrentFile //string        `bencode:"file"`
	//PieceList      []*Piece
	// md5sum for single files
	// files for multiple files
	// path for multiple files

	// Concatenation of all 20 byte SHA1
}

type TorrentFile struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"[]string"`
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
	data.PeerId = GenPeerId()
	// If Multiple Files
	if data.Info.Length == 0 {
		reader = bytes.NewReader(data.Info.FilesBytes)
		dec = bencode.NewDecoder(reader)
		dec.Decode(&data.Info.Files)
		data.Info.Length = data.Info.getTotalLength()
	} else {
		data.Info.SingleFile = true
	}
	data.Info.parsePieces()
	return &data, nil
}

func (info *TorrentInfo) getTotalLength() int64 {
	var total int64
	for _, file := range info.Files {
		total += file.Length
	}
	return total
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
