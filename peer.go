package main

import (
	"fmt"
)

var meta TorrentMeta

var blocks map[[20]byte]bool

func main() {
	meta, _ = ParseTorrent("tom.torrent")
	resp, err := GetTrackerResponse(meta)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
	//fmt.Println(meta.Info.Pieces[:20])
	//fmt.Println(len(meta.Info.Pieces) / 20)

}
