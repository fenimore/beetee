package main

import "fmt"

//import "net/http"

func main() {
	//d, _ := ParseTorrent("tom.torrent")
	d, _ := ParseTorrent("tom.torrent")
	//	fmt.Println(d.Info)
	//infoHash := sha1.Sum(d.InfoHash))
	//fmt.Println(d.InfoHash)
	//fmt.Println(d.InfoHex)
	//fmt.Println(d.InfoUrlEncoded)
	request := d.Announce + "?info_hash=" + d.InfoUrlEncoded
	fmt.Println(request)
	//resp, err := http.Get(d.Announce)
	//fmt.Println(resp, err)
}
