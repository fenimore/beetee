package main

import (
	"fmt"
	"github.com/jackpal/bencode-go"
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
	Announce     string
	AnounceList  []string
	CreationDate int
	Comment      string
	CreatedBy    string
	Encoding     string
	Info         map[string]interface{} //TorrentInfo
}

func (t *Torrent) String() string {
	return t.Announce
}

type TorrentInfo struct {
	PieceLength   int
	Pieces        string // Concatenation of all 20 byte SHA1
	Private       int    // 0 by default
	TorrentSingle        // This means TorrentInfo gets these fields?
	TorrentMultiple
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
	data := Torrent{}

	//data, err := bencode.Decode(reader)
	err = bencode.Unmarshal(reader, &data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(reflect.TypeOf(data))
	fmt.Println(data.CreatedBy)

}
