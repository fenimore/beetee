package main

import (
	"fmt"
	"github.com/zeebo/bencode"
	"os"
	"reflect"
)

//        self.announce = announce
// self.announce_list = announce_list
// self.comment = comment
// self.created_by = created_by
// self.created_at = created_at
// self.url_list = url_list
// self.raw_info = info
// self._parse_info(info)
type Torrent struct {
	//info ??
	Announce     string   `bencode:"announce"`
	AnounceList  []string `bencode:"announce-list"`
	CreatedBy    string   `bencode:"created by"`
	CreationDate int64    `bencode:"creation date"`
	Comment      string
	Encoding     string
	Info         TorrentInfo `bencode:"info"`
}

func (t *Torrent) String() string {
	return t.Announce
}

type TorrentInfo struct {
	PieceLength int64  `bencode:"piece length"`
	Private     int64  `bencode:"private"`
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
	Pieces      string `bencode:"pieces"`
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
	reader, err := os.Open("tom.torrent")
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
	fmt.Println(data.AnounceList)
}
