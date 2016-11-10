// beetee Parse package.
package main

import (
	"bytes"
	"fmt"
	//"github.com/zeebo/bencode"
	"crypto/sha1"
	"github.com/anacrolix/torrent/bencode"
	"os"
	//"strings"
)

// NOTE:
// (pieces/20)*piece length == length

type TorrentMeta struct {
	//info ??
	Announce string `bencode:"announce"`
	// Not tested announce-list yet
	AnounceList    []string `bencode:"announce-list"`
	CreatedBy      string   `bencode:"created by"`
	CreationDate   int64    `bencode:"creation date"`
	Comment        string
	Encoding       string        `bencode:"encoding"`
	UrlList        []string      `bencode:"url-list"`
	InfoBytes      bencode.Bytes `bencode:"info"`
	Info           TorrentInfo
	InfoHash       [20]byte
	InfoHex        string
	InfoUrlEncoded string
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
	// Parse the File
	dec := bencode.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return data, err
	}

	// Parse the Info Dictionary
	reader := bytes.NewReader(data.InfoBytes)
	dec = bencode.NewDecoder(reader)
	dec.Decode(&data.Info)

	// Compute the info_hash
	data.InfoHash = sha1.Sum(data.InfoBytes)
	data.InfoHex = fmt.Sprintf("%x", data.InfoHash)
	//data.InfoUrlEncoded = strings.ToUpper(data.InfoHex)
	data.InfoUrlEncoded = data.InfoHex
	var url string
	counter := 0
	for _, c := range data.InfoUrlEncoded {
		url += string(c)
		if counter == 1 {
			url += "%"
		}
		if counter == 1 {
			counter = 0
			continue
		}
		counter++
	}
	//fmt.Println(url)
	data.InfoUrlEncoded = url
	//fmt.Println(base64.URLEncoding.EncodeToString([]byte(string(data.InfoHash[:]))))
	return data, nil
}

func UrlEncode(hash [20]byte) string {
	var enc string
	for _, b := range hash {

	}

	return ""
}

// func main() {
//	data, _ := ParseTorrent("tom.torrent")
//	fmt.Println(data.Info.Pieces[20:40])
// }
