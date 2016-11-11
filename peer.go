package main

import (
	"fmt"
	"net"
)

type Peer struct {
	meta *TorrentMeta // Whence it came

	PeerId string `bencode:"peer id"` // Bencoding not being used
	Ip     string `bencode:"ip"`
	Port   uint16 `bencode:"port"`
	// After connection
	Shaken bool
	Conn   net.Conn
	Id     string
}

func ConnectToPeer(peer *Peer) (*Peer, error) {
	addr := fmt.Sprintf("%s:%d", peer.Ip, peer.Port)
	logger.Println("Connecting to Peer: ", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		debugger.Println(err)
	}

	hs := NewHandShake(peer.meta, conn)
	pId, err := hs.ShakeHands()
	if err != nil {
		debugger.Println(err)
	}
	peer.Id = pId
	peer.Shaken = true
	logger.Println("Connected to Peer: ", pId)

	return peer, nil

}
