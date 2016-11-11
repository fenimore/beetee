package main

import (
	//"bufio"
	"fmt"
	"net"
)

var meta TorrentMeta

var peerId string

var blocks map[[20]byte]bool

func main() {

	peerId = GenPeerId()

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

	//fmt.Println(resp.Complete, resp.Incomplete)
	//fmt.Println(len(resp.PeerList))
	//fmt.Println(resp.PeerList)

	/*TODO: Connect to Peer*/
	target := fmt.Sprintf("%s:%d", resp.PeerList[1].Ip, resp.PeerList[1].Port)
	fmt.Println("Connecting to ", target)
	conn, err := net.Dial("tcp", target)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	// status, err := bufio.NewReader(conn).Read()
	// if err != nil {
	//	fmt.Println(err)
	// }
	// fmt.Println(string(status))
}
