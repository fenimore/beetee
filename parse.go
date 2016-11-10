// beetee Parse package.
package main

import (
	"bytes"
	"fmt"
	//"github.com/zeebo/bencode"
	"github.com/anacrolix/torrent/bencode"
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
	Encoding     string        `bencode:"encoding"`
	UrlList      []string      `bencode:"url-list"`
	InfoBytes    bencode.Bytes `bencode:"info"`
	Info         TorrentInfo
	InfoHash     []byte
	//map[string]interface{} `bencode:"info"`
	//Info         TorrentInfo
	// TODO: Save info as bytes
	// and then hash them

}

func (t *TorrentMeta) String() string {
	return t.Announce
}

type TorrentInfo struct {
	Length      int64         `bencode:"length"`
	Name        string        `bencode:"name"`
	PieceLength int64         `bencode:"piece length"`
	Pieces      bencode.Bytes `bencode:"pieces"`
	Private     int64         `bencode:"private"`
	// md5sum for single files
	// files for multiple files
	// path for multiple files
	//Files       string `bencode:"file"`
	// Concatenation of all 20 byte SHA1
}

func ParseTorrent(file string) (TorrentMeta, error) {
	var data TorrentMeta
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
	reader := bytes.NewReader(data.InfoBytes)
	dec = bencode.NewDecoder(reader)
	dec.Decode(&data.Info)
	return data, nil
}

func main() {
	d, _ := ParseTorrent("tom.torrent")
	fmt.Println((len(d.Info.Pieces) / 20) * int(d.Info.PieceLength))
	fmt.Println(int(d.Info.Length))
}
