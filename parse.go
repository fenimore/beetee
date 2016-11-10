// beetee Parse package.
package main

import (
	"github.com/zeebo/bencode"
	"os"
)

// NOTE:
// (pieces/20)*piece length == length

type TorrentMeta struct {
	//info ??
	Announce string `bencode:"announce"`
	// Not tested announce-list yet
	AnounceList  []string `bencode:"announce-list"`
	CreatedBy    string   `bencode:"created by"`
	CreationDate int64    `bencode:"creation date"`
	Comment      string
	Encoding     string
	UrlList      []string               `bencode:"url-list"`
	Info         map[string]interface{} `bencode:"info"`
	//Info         TorrentInfo            `bencode:"info"`
	// TODO: Save info as bytes
	// and then hash them

}

func (t *TorrentMeta) String() string {
	return t.Announce
}

type TorrentInfo struct {
	Length      int64  `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int64  `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     int64  `bencode:"private"`
	// md5sum for single files
	// files for multiple files
	// path for multiple files
	//Files       string `bencode:"file"`
	// Concatenation of all 20 byte SHA1
}

type TorrentSingle struct {
	Name   string
	Length int
	Md5    string // a hex string
}

type TorrentMultiple struct {
	// Name
	// Files
	//     length md5 and path
}

func ParseTorrent(file string) (Torrent, error) {
	var data Torrent
	f, err := os.Open(file)
	if err != nil {
		return data, err
	}
	defer f.Close()
	dec := bencode.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}
