package main

import (
	"crypto/sha1"
	"time"
)

func Deluge() {
	pieceChan := make(chan *Piece, len(Pieces))
	recycleChan := make(chan *Piece)
	order := DecidePieceOrder()
	for _, idx := range order {
		pieceChan <- Pieces[idx]
	}
	debugger.Printf("Queue Filled")
	for _, peer := range Peers[:4] {
		debugger.Printf("Launch goroutine for peer")
		go HandlePeer(peer, pieceChan, recycleChan)

	}
	for {
		select {
		case recycle := <-recycleChan:
			pieceChan <- recycle
		}
	}
}

func HandlePeer(peer *Peer, pieces <-chan *Piece, recycle chan<- *Piece) {
	err := peer.ConnectPeer()
	if err != nil {
		debugger.Printf("Error Connected %s", err)
		return
	}
PeerLoop:
	for {
		if peer.choked {
			peer.sendStatusMessage(InterestedMsg)
		} else {
			peer.sendStatusMessage(-1) // keep alive
		}

		peer.choke.Wait() // if Choked, then Wait

		select {
		case <-peer.stopping:
			return
		default:
			// move along
		}
		piece := <-pieces
		if peer.bitfield == nil {
			continue
		}
		if !peer.bitfield[piece.index] {
			debugger.Printf("Peer %s doesn't have piece %d",
				peer.id, piece.index)
			recycle <- piece
			continue PeerLoop
		}
		//piece.pending.Add(1)
		peer.requestPiece(piece.index)

		for {
			select {
			case <-piece.success:
				piece.writeBlocks()
				continue PeerLoop
			case <-time.After(time.Second * 30):
				recycle <- piece
				debugger.Printf("Peer %s Loop Restarts", peer.id)
				continue PeerLoop
			}
		}
	}

}

func WaitForPieceToFill() {

}

// Flood is the run() of beetee.
func Flood() {
	order := DecidePieceOrder() // TODO: Rarest first?
	debugger.Println("Filling up QUeue")
	for _, idx := range order {
		PieceQueue <- Pieces[idx]
	}
	debugger.Println("QUeue Filled")
	// TODO: add queue for peers
	for _, peer := range Peers[:16] {
		err := peer.ConnectPeer()
		if err != nil {
			debugger.Printf("Error Connected to %s: %s", peer.addr, err)
			continue
		}
		//go peer.AskPeer()
		go peer.AskForDataFromPeer()
	}

}

func (peer *Peer) PeerManager() {
	for {
		select {
		case <-peer.stopping:
			debugger.Printf("Peer %s is closing", peer.id)
			return
		case <-peer.choking:
			if !peer.choked {
			}
		default:
			peer.choke.Wait()
		}
	}
}

func (peer *Peer) AskForDataFromPeer() {
	peer.sendStatusMessage(InterestedMsg)
	//peer.se
	for {
		if !peer.alive {
			return
		}
		//debugger.Printf("Peer %s Choked?", peer.id)
		peer.choke.Wait() // if Choked, then Wait
		//debugger.Printf("Peer %s Unchoke", peer.id)
		piece := <-PieceQueue
		if !peer.bitfield[piece.index] {
			PieceQueue <- piece
			continue
		}
		piece.pending.Add(1)
		peer.requestPiece(piece.index)
		//piece.timeout = time.Now()
		// Don't ask the same peer for too many pieces
		piece.pending.Wait()
	}
}

func (peer *Peer) AskPeer() {
	peer.sendStatusMessage(InterestedMsg)
	//peer.se
	for {
		if !peer.alive {
			return
		}

		//debugger.Printf("Peer %s Choked?", peer.id)
		peer.choke.Wait() // if Choked, then Wait
		//debugger.Printf("Peer %s Unchoke", peer.id)
		piece := <-PieceQueue
		if !peer.bitfield[piece.index] {
			PieceQueue <- piece
			continue
		}
		piece.pending.Add(1)
		//peer.requestPiece(piece.index)
		//piece.timeout = time.Now()
		go piece.PieceManager(peer)
		// Don't ask the same peer for too many pieces
		piece.pending.Wait()
	}
}

// PieceManager must be run for every piece..
// NOTE: Should I request pieces here?
// TODO: add pending here.
// FIXME: hash duplicates?
func (piece *Piece) PieceManager(peer *Peer) {
	// TODO: Set as global?
	debugger.Printf("From %s downloading %d", peer.id, piece.index)
	peer.requestPiece(piece.index)
	numberOfBlocks := int(Torrent.Info.PieceLength) / blocksize
	var blockCount int
ForLoop:
	for {
		select {
		case block := <-piece.chanBlocks:
			// NOTE: copy block to piece data
			copy(piece.data[int(block.offset):int(
				block.offset)+blocksize],
				block.data)
			if blockCount == numberOfBlocks {
				// TODO: Check for hash here or elsewhere?
				if piece.hash != sha1.Sum(piece.data) {
					piece.data = nil
					piece.data = make([]byte,
						piece.size)
					debugger.Printf("Invalid Hash of  Piece %d",
						piece.index)
					PieceQueue <- piece // NOTE Return to queue
					piece.pending.Done()
					return
				}
				break ForLoop
			} else if piece.index == len(Pieces)-1 {
				// Last Piece could have fewer blocks // FIXME REally??
				if piece.hash == sha1.Sum(piece.data) {
					break ForLoop
				}
			}
			continue ForLoop // NOTE Not enough block
		case <-peer.stopping:
			piece.data = nil
			piece.data = make([]byte,
				piece.size)
			debugger.Printf(
				"Peer %s Stops: Download incompletes %d",
				peer.id, piece.index)
			PieceQueue <- piece // NOTE Return
			piece.pending.Done()
			return
		case <-peer.choking:
			piece.data = nil
			piece.data = make([]byte,
				piece.size)
			debugger.Printf(
				"Peer %s Chokes: Download incompletes %d",
				peer.id, piece.index)
			PieceQueue <- piece // NOTE Return
			piece.pending.Done()
			return
		case <-time.After(time.Second * 30):
			debugger.Printf("Piece %d timeout", piece.index)
			PieceQueue <- piece
			piece.pending.Done()
			return
		default:
			continue ForLoop
		}
	}
	piece.verified = true
	logger.Printf("Piece at %d is successfully written",
		piece.index)
	ioChan <- piece
	piece.pending.Done()
	return

}

// of pieces, according to the rarest first
func DecidePieceOrder() []int {
	order := make([]int, 0, len(Pieces))
	for i := 0; i < len(Pieces); i++ {
		if !Pieces[i].verified {
			order = append(order, i)
		}
	}
	return order
}
