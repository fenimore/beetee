package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

var ( // NOTE Global Important Variables
	Torrent *TorrentMeta
	Peers   []*Peer
	Pieces  []*Piece
	PeerId  [20]byte
	// Channels
	PieceQueue  chan *Piece
	PeerQueue   chan *Peer
	ActivePeers []*Peer
	// Status ongoing
	Left       int
	Uploaded   int
	AliveDelta int // TODO:
	MaxPeers   int
	// Loggers
	debugger *log.Logger
	logger   *log.Logger
	// WaitGroup
	completionSync sync.WaitGroup
	writeSync      sync.WaitGroup
)

//var continWG sync.WaitGro[up

var (
	peerId [20]byte
	blocks map[[20]byte]bool

	wg sync.WaitGroup
)

func main() {
	// Exit Write
	c := make(chan os.Signal, 1) // SIGINT
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		Torrent.Info.WriteData()
		os.Exit(1)
	}()
	// Debug and Error variables
	var err error
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)

	// My Peer Id, unique...
	PeerId = GenPeerId()

	/* Parse Torrent*/
	Torrent, err = ParseTorrent("torrents/tom.torrent")
	if err != nil {
		debugger.Println(err)
	}
	debugger.Println("Length: ", Torrent.Info.Length)
	debugger.Println("Piece Length: ", Torrent.Info.PieceLength)
	debugger.Println("Piece Len: ", len(Torrent.Info.Pieces))
	debugger.Println("Pieces count: ", len(Pieces))
	/*Parse Tracker Response*/
	_, err = GetTrackerResponse(Torrent)
	if err != nil {
		debugger.Println(err)
	}
	/* What next? */
	PieceQueue = make(chan *Piece)
	PeerQueue = make(chan *Peer, MaxPeers)
	MaxPeers = len(Peers) / 2
	writeSync.Add(1)
	go Torrent.Info.ContinuousWrite()
	go Flood()
	writeSync.Wait()
	//completionSync.Wait()
	//err = Torrent.Info.WriteData()
	//if err != nil {
	//		logger.Printf("Problem writing data %s", err)
	//		os.Exit(0)
	//	} else {
	//		logger.Printf("Wrote Data NP")
	//		os.Exit(0)
	//	}
}
