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
	fmt.Println(len(resp.Peers))
	fmt.Println(len(resp.Peers) / 6)
	fmt.Println(resp.Peers[4])

	//p := resp.Peers[4:]
	//ip := net.IPv4(p[0], p[1], p[2], p[3])
	//port := (uint16(p[4]) << 8) | uint16(p[5])
	//fmt.Println(port)
	//fmt.Println(ip.String())

	/*TODO: Connect to Peer*/

}
