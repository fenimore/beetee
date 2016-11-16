package main

import (
	"log"
	"os"
	"sync"
)

var ( // NOTE Global Important Variables
	Torrent *TorrentMeta
	Peers   []*Peer
	Pieces  []*Piece
	PeerId  [20]byte
	// Channels
	PieceQueue chan *Piece
	// Status ongoing
	Left       int
	Uploaded   int
	AliveDelta int // TODO:
	// Loggers
	debugger *log.Logger
	logger   *log.Logger
	// WaitGroup
	completionSync sync.WaitGroup
)

var (
	peerId [20]byte
	blocks map[[20]byte]bool

	wg sync.WaitGroup
)

func main() {
	var err error

	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)

	PeerId = GenPeerId()

	/* Parse Torrent*/
	Torrent, err = ParseTorrent("torrents/tom.torrent")

	//debugger.Println(meta.Info)

	if err != nil {
		debugger.Println(err)
		//fmt.Println(err)
	}
	debugger.Println("Length: ", Torrent.Info.Length)
	debugger.Println("Piece Length: ", Torrent.Info.PieceLength)
	debugger.Println("Piece Len: ", len(Torrent.Info.Pieces))
	debugger.Println(len(Pieces))
	/*Parse Tracker Response*/
	_, err = GetTrackerResponse(Torrent)
	if err != nil {
		debugger.Println(err)
	}
	/* What next? */
	PieceQueue = make(chan *Piece)
	Flood()
	completionSync.Wait()
	err = Torrent.Info.WriteData()
	if err != nil {
		logger.Printf("Problem writing data %s", err)
		os.Exit(0)
	} else {
		logger.Printf("Wrote Data NP")
		os.Exit(0)
	}
	// peer := Peers[1]
	// err = peer.ListenToPeer()
	// if err != nil {
	//	debugger.Println("Error connection", err)
	// } else {
	//	// is connectedy
	//	peer.sendStatusMessage(InterestedMsg)
	//	peer.ChokeWg.Wait()
	//	//peer.requestBlock(78, 0)
	//	completionSync.Add(len(Pieces) - 1)
	//	peer.requestAllPieces()
	//	completionSync.Wait()

	//	err = Torrent.Info.WriteData()
	//	if err != nil {
	//		logger.Printf("Problem writing data %s", err)
	//		os.Exit(0)
	//	} else {
	//		logger.Printf("Wrote Data NP")
	//		os.Exit(0)
	//	}

	// }
	/* TODO: Request Blocks */

}
