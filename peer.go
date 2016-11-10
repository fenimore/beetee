package main

import (
	"fmt"
	//"net"
)

var meta TorrentMeta

var blocks map[[20]byte]bool

func main() {
	/* Parse Torrent*/
	meta, err := ParseTorrent("ubuntu.torrent")
	if err != nil {
		fmt.Println(err)
	}

	/*Parse Tracker Response*/
	resp, err := GetTrackerResponse(meta)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: NOT really working
	fmt.Println(resp.Complete, resp.Incomplete)
	fmt.Println(len(resp.PeerList))
	fmt.Println(resp.PeerList)
	/*TODO: Connect to Peer*/

}
