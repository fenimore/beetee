package main

import (
	"log"
	"os"
)

var (
	//meta   TorrentMeta
	peerId [20]byte
	blocks map[[20]byte]bool

	debugger *log.Logger
	logger   *log.Logger
)

func main() {
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)

	peerId = GenPeerId()

	/* Parse Torrent*/
	meta, err := ParseTorrent("tom.torrent")
	if err != nil {
		debugger.Println(err)
		//fmt.Println(err)
	}

	/*Parse Tracker Response*/
	resp, err := GetTrackerResponse(meta)
	if err != nil {
		debugger.Println(err)
	}

	// /*TODO: Connect to Peer*/
	peer, err := ConnectToPeer(resp.PeerList[1])
	if err != nil {
		debugger.Println(err)
	}
	logger.Println(peer.Id)
}
