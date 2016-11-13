package main

import (
	"log"
	"os"
	"sync"
	"time"
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
	//	logger.Println([]byte{(uint8)(0),
	//		(uint8)(0), (uint8)(0), (uint8)(1)})

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

	//logger.Println("Pieces:\n\n", string(meta.Info.Pieces))
	//return
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
	/* Listen to Peer */
	logger.Println(peer.Id)
	go peer.ListenToPeer()

	/* Tell Peer I'm interested */
	time.Sleep(40000)
	err = peer.sendStatusMessage(InterestedMsg)
	if err != nil {
		debugger.Println("Status Error: ", err)
	}
	wg.Add(1)

	/* TODO: Request Blocks */
	wg.Wait()
}
