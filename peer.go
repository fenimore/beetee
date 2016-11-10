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

	//fmt.Println(string(resp.Peers))
	//fmt.Println(resp.Peers)
	//fmt.Println(reflect.TypeOf(resp.Peers))
	//fmt.Println(resp.Peers[:6])
	// TODO: NOT really working
	fmt.Println(resp.Peers[:2])
	p := resp.Peers[3:7]
	ip := net.IPv4(p[0], p[1], p[2], p[3])
	fmt.Println(ip.String())
	p = resp.Peers[6:12]
	ip = net.IPv4(p[0], p[1], p[2], p[3])
	fmt.Println(ip.String())

	//fmt.Println(resp.Complete, resp.Incomplete)

	/*TODO: Connect to Peer*/

}

// Convert uint to net.IP
func inet_ntoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}
