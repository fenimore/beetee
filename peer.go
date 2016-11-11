package main

import (
	"fmt"
	"net"
)

var meta TorrentMeta

var peerId [20]byte

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

	// /*TODO: Connect to Peer*/
	peer := fmt.Sprintf("%s:%d", resp.PeerList[1].Ip, resp.PeerList[1].Port)
	fmt.Println("Connecting to ", peer)
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(WriteHandShake(meta)))	conn.Write(WriteHandShake(meta))
	hs := NewHandShake(meta, conn)
	s, e := hs.ShakeHands()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(s)
}
