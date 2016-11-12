package main

import (
	"log"
	"os"
	"sync"
)

var (
	//meta   TorrentMeta
	peerId [20]byte
	blocks map[[20]byte]bool

	debugger *log.Logger
	logger   *log.Logger

	wg sync.WaitGroup
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

	debugger.Println("Length: ", meta.Info.Length)
	debugger.Println("Piece Length: ", meta.Info.PieceLength)
	debugger.Println("Piece Len: ", len(meta.Info.Pieces))
	//logger.Println("Pieces: ", meta.Info.Pieces)

	/*Parse Tracker Response*/
	resp, err := GetTrackerResponse(meta)
	if err != nil {
		debugger.Println(err)
	}

	/*Connect to Peer*/
	peer := resp.PeerList[1]
	err = peer.ConnectToPeer()
	if err != nil {
		debugger.Println(err)
	}

	logger.Println(peer.Id)
	go peer.ListenToPeer()
	wg.Add(1)

	/* TODO: Request Blocks */
	wg.Wait()
}
