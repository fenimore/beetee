package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
)

const port = 6881 // TODO

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
	// Loggers
	debugger *log.Logger
	logger   *log.Logger
	// WaitGroup
	completionSync sync.WaitGroup
	writeSync      sync.WaitGroup
	queueSync      sync.WaitGroup
	// Mutex
)

var (
	pieceChan chan *Piece
	peerChan  chan *Peer
	ioChan    chan *Piece
)

func main() {
	/* Get Arguments */
	torrentFile := flag.String("file", "torrents/tom.torrent", "path to torrent file")
	flag.Usage = func() {
		fmt.Println("beetee, commandline torrent application. Usage:")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	/* Exit on CTRL C */
	c := make(chan os.Signal, 1) // SIGINT
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		//Torrent.Info.WriteData()
		debugger.Println("Good Bye!")
		os.Exit(2)
	}()

	/* TODO: Start server on Port */
	// server

	/* Debug and Error variables */
	var err error
	debugger = log.New(os.Stdout, "DEBUG: ",
		log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ",
		log.Ltime|log.Lshortfile)

	/* My Peer Id TODO unique */
	PeerId = GenPeerId()

	/* Start Listening */
	server := NewServer()
	defer server.listener.Close()
	go server.Listen()

	/* Parse Torrent*/
	// NOTE: Sets Piece
	Torrent, err = ParseTorrent(*torrentFile)
	if err != nil {
		debugger.Println(err)
	}

	debugger.Println("Length: ", Torrent.Info.Length)
	debugger.Println("Piece Length: ",
		Torrent.Info.PieceLength)
	debugger.Println("Piece Len: ",
		len(Torrent.Info.Pieces))
	debugger.Println("Pieces count: ", len(Pieces))

	/*Parse Tracker Response*/
	// NOTE: Sets Peer
	_, err = GetTrackerResponse(Torrent)
	if err != nil {
		debugger.Println(err)
	}

	/* Start Client */
	//PieceQueue = make(chan *Piece, len(Pieces))
	PeerQueue = make(chan *Peer)
	ioChan = make(chan *Piece, len(Pieces))
	//pieces := make(chan *Piece, len(Pieces))
	//peers := make(chan *Peer)
	go Torrent.Info.FileWrite()
	Deluge()
	//go FileWrite()
	//Flood() //pieces, peers)
	writeSync.Add(1)
	writeSync.Wait()
}
