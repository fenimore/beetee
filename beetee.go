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
	BLOCKSIZE           = 16384
	PORT                = 6939 // TODO
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
	// Loggers
	debugger *log.Logger
	logger   *log.Logger
	// WaitGroup
	writeSync     sync.WaitGroup
	resetPeerList *sync.WaitGroup
	// Protocol Values
	pstr    = []byte("BitTorrent protocol")
	pstrlen = byte(19)
	d       *Download
)

func main() {
	d = &Download{
		Peers:  make([]*Peer, 0),
		Pieces: make([]*Piece, 0),
		Left:   0, // TODO: set according to partial download
	}

	/* Get Arguments */
	torrentFile := flag.String("path", "torrents/tom.torrent", "path to torrent file")
	seedTorrent := flag.Bool("seed", false, "keep running after download completes")
	maxPeers := flag.Int("peers", 40, "max peer connections")
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
		for _, p := range d.Pieces {
			if !p.verified {
				debugger.Println("Piece Not found:", p.index)
			}
		}

		curSize, err := checkFileSize(d.Torrent.Info.Name)
		if err != nil {
			debugger.Println(err)
		}
		debugger.Printf("File is %d bytes, out of Length: %d",
			curSize, d.Torrent.Info.Length)
		debugger.Println("Good Bye!")
		os.Exit(2)
	}()

	/* Debug and Error variables */
	var err error
	debugger = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lshortfile)
	logger = log.New(os.Stdout, "LOG: ", log.Ltime|log.Lshortfile)

	/* Start Listening */
	// TODO: use leechers
	leechers := Serve(PORT, make(chan bool))

	/* Parse Torrent*/
	// NOTE: Sets Piece
	// NOTE: Sets Peer list
	d.Torrent, err = ParseTorrent(*torrentFile)
	if err != nil {
		debugger.Println(err)
	}

	debugger.Println("File Length: ", d.Torrent.Info.Length)
	debugger.Println("Piece Length: ", d.Torrent.Info.PieceLength)
	debugger.Println(d.Torrent.Info.lastPieceSize())
	debugger.Println("len(info.Pieces) // bytes: ", len(d.Torrent.Info.Pieces))
	debugger.Println("len(Pieces) // pieces: ", len(d.Pieces))
	debugger.Println("len(Peers) // peers: ", len(d.Peers))
	debugger.Println("Port listening on", PORT)

	/*Parse Tracker Response*/
	err = GetTrackerResponse(d.Torrent)
	if err != nil {
		debugger.Println(err)
	}

	waiting := make(chan *Peer, *maxPeers)
	ready := make(chan *Peer) // Unchoked
	choked := make(chan *Peer)
	leeching := make(chan *Peer)
	disconnected := make(chan *Peer)

	pieceNext := FillPieceOrder()
	// NOTE: Backwards for testing last piece.
	//pieceNext := Backwards()

	// Start writing to disk
	diskIO, closeIO := spawnFileWriter(d.Torrent.Info.Name,
		d.Torrent.Info.SingleFile, d.Torrent.Info.Files, pieceNext)
	go func() {
		for _, peer := range d.Peers {
			waiting <- peer
		}
	}()

	// Peers not yet connected/handshaken
	go func() {
		for peer := range waiting {
			go peer.spawnPeerHandShake(waiting, choked, ready)
		}
	}()

	// Peers having been handshaked
	go func() {
		for peer := range choked {
			peer.spawnPeerHandler(waiting, choked, ready, disconnected)
			peer.spawnPeerReader()
			// if peer has what I want TODO:
			// XOR my bitmap with theirse
			peer.out <- StatusMessage(InterestedMsg)
		}
	}()

	go func() {
		for peer := range leechers {
			debugger.Println("New Leecher", peer.id)
			// leechers have already been handshaken
			peer.spawnPeerReader()
			peer.spawnPeerHandler(waiting, choked, ready, disconnected)
			leeching <- peer
		}
	}()

	go func() {
		for peer := range ready {
			go func(p *Peer) {

				for {
					if string(p.id) == "-LT1200-YpsXDHDalf2z" {
						return
					}
					piece := <-pieceNext
					logger.Printf("Requesting Piece %d From %s",
						piece.index, p.id)
					if !p.bitmap[piece.index] {
						pieceNext <- piece
						continue
					}
					msgs := requestPiece(piece.index)
					for _, msg := range msgs {
						p.out <- msg
					}
					select {
					// TODO: stop when peer closes
					case <-piece.success:
						diskIO <- piece
					case <-time.After(30 * time.Second):
						logger.Println("TimeOut Pieces", piece.index)
						pieceNext <- piece
						close(p.halt)
						return
					}
				}
			}(peer)
		}
	}()

	if *seedTorrent {
		writeSync.Add(1)
	}

	/* Finish or Cancel downloading */
	writeSync.Wait()
	close(closeIO)
	for _, p := range d.Pieces {
		if !p.verified {
			debugger.Println("Piece Not found:", p.index)
		}
	}
	curSize, err := checkFileSize(d.Torrent.Info.Name)
	if err != nil {
		debugger.Println(err)
	}
	debugger.Printf("File is %d bytes, out of Length: %d",
		curSize, d.Torrent.Info.Length)
	debugger.Println("Good Bye!")

}

// FillPieceOrder is a placeholder for rarest first
// algorithm. It simply puts all desired pieces into
// a channel for request.
func FillPieceOrder() chan *Piece {
	out := make(chan *Piece, len(d.Pieces))
	for i := 0; i < len(d.Pieces); i++ {
		out <- d.Pieces[i]
	}
	return out
}

// Backwards will ask for pieces in reverse order.
// useful for debugging the last piece problem.
func Backwards() chan *Piece {
	out := make(chan *Piece, len(d.Pieces))
	for i := len(d.Pieces) - 1; i >= len(d.Pieces)-2; i-- {
		out <- d.Pieces[i]
	}
	return out
}
