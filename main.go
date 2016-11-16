package main

import (
	"log"
	"os"
	"sync"
)

var (
	//meta   TorrentMeta
	manager Manager
	peerId  [20]byte
	blocks  map[[20]byte]bool

	debugger *log.Logger
	logger   *log.Logger

	wg sync.WaitGroup
)

func main() {
	manager := new(Manager)
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)
	// logger.Println([]byte{(uint8)(0),
	//	(uint8)(0), (uint8)(1), (uint8)(3)})
	// return

	peerId = GenPeerId()

	/* Parse Torrent*/
	meta, err := ParseTorrent("arch.torrent")
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

	/*Connect to Peer*/
	// for idx, p := range resp.PeerList {
	//	if idx > 2 {
	//		break
	//	}
	//	go p.ListenToPeer()
	// }
	/* Tell Peer I'm interested */
	peers := resp.PeerList
	manager.peers = peers
	for _, val := range peers {
		logger.Println(val)
	}
	//go resp.PeerList[3].ListenToPeer()
	// requestAllPieces
	//resp.PeerList[8].requestAllPieces()
	wg.Add(1)

	/* TODO: Request Blocks */
	wg.Wait()
}
