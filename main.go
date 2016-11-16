package main

import (
	"log"
	"os"
	"sync"
)

var (
	//meta   TorrentMeta
	manager *Manager
	peerId  [20]byte
	blocks  map[[20]byte]bool

	debugger *log.Logger
	logger   *log.Logger

	wg sync.WaitGroup
)

func main() {
	manager = new(Manager)
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)
	// logger.Println([]byte{(uint8)(0),
	//	(uint8)(0), (uint8)(1), (uint8)(3)})
	// return

	peerId = GenPeerId()

	/* Parse Torrent*/
	meta, err := ParseTorrent("ubuntu.torrent")
	if err != nil {
		debugger.Println(err)
		//fmt.Println(err)
	}
	debugger.Println("Length: ", meta.Info.Length)
	debugger.Println("Piece Length: ", meta.Info.PieceLength)
	debugger.Println("Piece Len: ", len(meta.Info.Pieces))
	// debugger.Println("PieceList:", meta.Info.PieceList)
	// debugger.Println("Pieces:\n\n", meta.Info.Pieces)

	/*Parse Tracker Response*/
	resp, err := GetTrackerResponse(meta)
	if err != nil {
		debugger.Println(err)
	}

	/* Set Madnager traits*/
	manager.peers = resp.PeerList
	manager.torrent = &meta
	manager.left = int(meta.Info.Length)
	manager.pieceQ = make(chan int)
	manager.pieceStatus = make(map[int]int)
	/* Launch Manager */
	manager.Flood()

	wg.Add(1)

	/* TODO: Request Blocks */
	wg.Wait()
}
