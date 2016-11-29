package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

const port = 6881 // TODO

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
		file, _ := os.Open(Torrent.Info.Name)
		fi, _ := file.Stat()
		debugger.Printf("File is %d bytes, out of Length: %d",
			fi.Size(), Torrent.Info.Length)
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
	Torrent, err = ParseTorrent(*torrentFile)
	if err != nil {
		debugger.Println(err)
	}

	debugger.Println("File Length: ", Torrent.Info.Length)
	debugger.Println("Piece Length: ", Torrent.Info.PieceLength)
	debugger.Println("len(info.Pieces) // bytes: ", len(Torrent.Info.Pieces))
	debugger.Println("len(Pieces) // pieces: ", len(Pieces))

	/*Parse Tracker Response*/
	tr, err := GetTrackerResponse(Torrent)
	if err != nil {
		debugger.Println(err)
	}

	// Get Peers
	Peers = ParsePeers(tr)

	/* Start Client */
	PieceQueue = make(chan *Piece, len(Pieces))
	msgChan = make(chan []byte)
	PeerQueue = make(chan *Peer)
	ioChan = make(chan *Piece, len(Pieces))

	for _, peer := range Peers {
		if peer.addr == "207.251.103.46:6882" {
			continue
		}
		conn, err := peer.Connect()
		if err != nil {
			debugger.Printf("Connect error %s", err)
			continue
		}
		err = peer.HandShake(conn, Torrent)
		if err != nil {
			debugger.Printf("Handshake error %s", err)
			continue
		}

		peer.chokeWg.Add(1)
		go func(p *Peer, c net.Conn) {
			for {
				err = p.DecodeMessages(c)
				if err != nil {
					debugger.Printf("Error Decoding: %s", err)
					break
				}
			}
			c.Close()
			debugger.Printf("Connection Closing to Peer %s", p.id)
		}(peer, conn)

		go func(p *Peer, c net.Conn) {
			// TODO: Send interested if bitfield has been received
			debugger.Println("Sending Interested")
			c.Write(StatusMessage(InterestedMsg))
			//p.chokeWg.Wait()
			debugger.Println("Recv Unchoke Message, Stop waiting")
			for idx, val := range Pieces {
				requests := requestPiece(idx)
				val.pending.Add(1)
				for _, r := range requests {
					debugger.Printf("Request Away: %s", r)
					c.Write(r)
				}
				val.pending.Wait()

			}
		}(peer, conn)
	}

	writeSync.Add(1)
	writeSync.Wait()
}
