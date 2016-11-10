package main

import (
	"fmt"
	"net"
)

var meta TorrentMeta

var blocks map[[20]byte]bool

func main() {
	/* Parse Torrent*/
	meta, err := ParseTorrent("tom.torrent")
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

	prefix := resp.Peers[:3] // first three (could be more
	for _, x := range prefix {
		fmt.Print(string(x))
	}
	fmt.Println()

	p := resp.Peers[:9]
	ip := net.IPv4(p[3], p[4], p[5], p[6])
	fmt.Println(ip.String())
	peerPort := uint16(p[7]) << 8
	peerPort = peerPort | uint16(p[8])
	fmt.Println(peerPort)
	p = resp.Peers[7:]
	ip = net.IPv4(p[0], p[1], p[2], p[3])
	fmt.Println(ip.String())

	/*TODO: Connect to Peer*/

}
