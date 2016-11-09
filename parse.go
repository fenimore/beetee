package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
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
	CreateBy     string
	Encoding     string
	Info         TorrentInfo
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

// Parsing!!!!!!!!!!!

//Strings
// Example: 4: spam represents the string "spam"
// Parses the byte slice, returns the string found
// and returns the remaining bytes to parse
func parseString(b []byte) (string, []byte, error) {
	var s string
	var numAsString string
	// Get the number of bytes that are the strings
	for _, val := range b {
		if string(val) != ":" {
			numAsString += string(val)
			continue
		}
		break
	}
	amt, err := strconv.Atoi(numAsString)
	if err != nil {
		return "", b, err
	}
	// TODO: get string Length
	fmt.Println(amt)

	return s, b, nil
}

//Integers
//    Example: i3e represents the integer "3"

//Lists
//Example: l4:spam4:eggse represents the list of two strings: [ "spam", "eggs" ]
//     Example: le represents an empty list: []

//Dictionaries
//Example: d3:cow3:moo4:spam4:eggse represents the dictionary { "cow" => "moo", "spam" => "eggs" }
//Example: d4:spaml1:a1:bee represents the dictionary { "spam" => [ "a", "b" ] }
//Example: d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee represents { "publisher" => "bob", "publisher-webpage" => "www.example.com", "publisher.location" => "home" }
//Example: de represents an empty dictionary {}

func main() {

	b, err := ioutil.ReadFile("tom.torrent")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}
