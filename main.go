package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	PORT                = 6881 // TODO
	FILE_WRITER_BUFSIZE = 25
)

type Download struct {
	Torrent *TorrentMeta
	Peers   []*Peer
	Pieces  []*Piece

	bitfield []byte
	Left     int
	Uploaded int
}

var ( // NOTE Global Important Variables
	// Channels
	PieceQueue chan *Piece
	PeerQueue  chan *Peer
	// Loggers
	debugger *log.Logger
	logger   *log.Logger
	// WaitGroup
	writeSync sync.WaitGroup
	// Protocol Values
	pstr    = []byte("BitTorrent protocol")
	pstrlen = byte(19)
	d       *Download

	pieceChan chan *Piece
	peerChan  chan *Peer
	ioChan    chan *Piece
)

func spawnFileWriter(f *os.File) (chan *Piece, chan struct{}) {
	in := make(chan *Piece, FILE_WRITER_BUFSIZE)
	close := make(chan struct{})
	go func() {
		for {
			select {
			case piece := <-in:
				logger.Printf("Writing Data to piece %d", piece.index)
				f.WriteAt(piece.data, int64(piece.index)*piece.size)
			case <-close:
				f.Close()
			}
		}
	}()
	return in, close
}

func main() {
	d = &Download{
		Peers:  make([]*Peer, 0),
		Pieces: make([]*Piece, 0),
		Left:   0, // TODO: set according to partial download
	}

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
		file, _ := os.Open(d.Torrent.Info.Name)
		fi, _ := file.Stat()
		debugger.Printf("File is %d bytes, out of Length: %d",
			fi.Size(), d.Torrent.Info.Length)
		debugger.Println("Good Bye!")
		os.Exit(2)
	}()

	/* Debug and Error variables */
	var err error
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)

	/* Start Listening */
	// TODO:

	/* Parse Torrent*/
	// NOTE: Sets Piece
	d.Torrent, err = ParseTorrent(*torrentFile)
	if err != nil {
		debugger.Println(err)
	}

	debugger.Println("File Length: ", d.Torrent.Info.Length)
	debugger.Println("Piece Length: ", d.Torrent.Info.PieceLength)
	debugger.Println("len(info.Pieces) // bytes: ", len(d.Torrent.Info.Pieces))
	debugger.Println("len(Pieces) // pieces: ", len(d.Pieces))

	/*Parse Tracker Response*/
	tr, err := GetTrackerResponse(d.Torrent)
	if err != nil {
		debugger.Println(err)
	}

	// Get Peers
	d.Peers = ParsePeers(tr)

	file, err := os.Create(d.Torrent.Info.Name)
	if err != nil {
		debugger.Println("Unable to create file")
	}
	diskIO, _ := spawnFileWriter(file)
	peerChannels := make(map[*Peer]PeerChannels)
	connPeers := make(chan *Peer)
	readyPeers := make(chan *Peer)
	for _, peer := range d.Peers {
		in := make(chan []byte)
		out, halt := peer.spawnPeerHandler(in, d, connPeers, readyPeers)
		peerChannels[peer] = PeerChannels{in: in, out: out, halt: halt}

	}
	go func() {
		for {
			peer := <-connPeers
			channels := peerChannels[peer]
			channels.in <- StatusMessage(InterestedMsg)
		}
	}()

	go func() {
		for {
			peer := <-readyPeers
			channels := peerChannels[peer]
			for i := 0; i < len(d.Pieces); i++ {
				msgs := requestPiece(i)
				for _, msg := range msgs {
					channels.in <- msg
				}
				select {
				case <-d.Pieces[i].success:
					diskIO <- d.Pieces[i]
					continue
				case <-time.After(30 * time.Second):
					continue
				}
			}

		}
	}()

	//	for _, peer := range d.Peers {
	// msgs := requestPiece(idx)
	// for _, msg := range msgs {
	//	channels.in <- msg
	// }
	// select {
	// case <-d.Pieces[idx].success:
	//	continue
	// case <-time.After(30 * time.Second):
	//	continue
	// }
	//	}

	writeSync.Add(1)
	writeSync.Wait()
}

type PeerChannels struct {
	in   chan []byte
	out  chan []byte
	halt chan []byte
}
