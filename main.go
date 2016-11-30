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
	diskIO, closeIO := spawnFileWriter(file)

	waiting := make(chan *Peer)
	ready := make(chan *Peer) // Unchoked
	choked := make(chan *Peer)
	disconnected := make(chan *Peer)

	pieceNext := FillPieceOrder()
	debugger.Println(len(pieceNext))

	go func() {
		for _, peer := range d.Peers[:] {
			waiting <- peer
		}
	}()

	go func() {
		for {
			peer := <-waiting
			go peer.spawnPeerHandShake(waiting, choked, ready)
		}
	}()

	// Open communication
	go func() {
		for {

		}

	}()

	go func() {
		for {
			peer := <-choked
			peer.spawnPeerHandler(waiting, choked, ready, disconnected)
			peer.spawnPeerReader()
			// if peer has what I want TODO:
			// XOR my bitmap with theirse
			peer.out <- StatusMessage(InterestedMsg)
		}
	}()

	go func() {
		for {
			peer := <-ready
			logger.Println("New Peer U", peer.id)
			go func(p *Peer) {

				for {
					piece := <-pieceNext
					logger.Printf("Requesting Piece %d From %s",
						piece.index, peer.id)
					msgs := requestPiece(piece.index)
					for _, msg := range msgs {
						peer.out <- msg
					}
					select {
					// TODO: stop when peer closes
					case <-piece.success:
						logger.Println("Wrote Piece:", piece.index)
						diskIO <- piece
					case <-time.After(30 * time.Second):
						logger.Println("TimeOut Pieces", piece.index)
						pieceNext <- piece
					}
				}
			}(peer)
		}
	}()
	writeSync.Add(1)
	writeSync.Wait()
	close(closeIO)
}

func FillPieceOrder() chan *Piece {
	out := make(chan *Piece, len(d.Pieces))
	for i := 0; i < len(d.Pieces); i++ {
		out <- d.Pieces[i]
	}
	return out
}
