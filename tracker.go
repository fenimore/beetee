package main

import "fmt"

//import "net/http"
//import "crypto/sha1"

func main() {
	d, _ := ParseTorrent("tom.torrent")
	fmt.Println(d.Info)
	//	fmt.Println(d.Info)
	//infoHash := sha1.Sum(d.InfoHash))
	//fmt.Println(infoHash)
	//request := d.Announce + "?info_hash=" + infoHash
	//resp, err := http.Get(d.Announce)
	//fmt.Println(resp, err)
}
