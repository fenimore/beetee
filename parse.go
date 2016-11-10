package main

import (
	"fmt"
	"github.com/zeebo/bencode"
	"os"
	"reflect"
)

// NOTE:
// (pieces/20)*piece length == length

type Torrent struct {
	//info ??
	Announce     string   `bencode:"announce"`
	AnounceList  []string `bencode:"announce-list"`
	CreatedBy    string   `bencode:"created by"`
	CreationDate int64    `bencode:"creation date"`
	Comment      string
	Encoding     string
	UrlList      []string    `bencode:"url-list"`
	Info         TorrentInfo `bencode:"info"`
}

func (t *Torrent) String() string {
	return t.Announce
}

type TorrentInfo struct {
	Length      int64  `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int64  `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     int64  `bencode:"private"`

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

func main() {

	//b, err := ioutil.ReadFile("tom.torrent")
	reader, err := os.Open("archlinux.torrent")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(b))
	var data Torrent

	//data, err := bencode.Decode(reader)
	dec := bencode.NewDecoder(reader)
	err = dec.Decode(&data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(reflect.TypeOf(data.Info))
	fmt.Println(data.UrlList)
}
