package main

import (
	"time"
)

// Flood is the run() of beetee.
func Flood() {
	// TODO: add queue for peers
	for _, peer := range Peers[:] {
		err := peer.ConnectPeer()
		if err != nil {
			debugger.Printf("Error Connected to %s: %s", peer.addr, err)
			continue
		}
		go peer.AskPeer()
	}
	order := DecidePieceOrder() // TODO: Rarest first?
	for _, idx := range order {
		PieceQueue <- Pieces[idx]
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

func (peer *Peer) AskPeer() {
	peer.sendStatusMessage(InterestedMsg)
	//peer.se
	for {
		if !peer.alive {
			return
		}
		if peer.choked {
			continue
		}
		//peer.choke.Wait() // if Choked, then Wait
		piece := <-PieceQueue
		if !peer.bitfield[piece.index] {
			PieceQueue <- piece
			continue
		}
		piece.pending.Add(1)
		peer.requestPiece(piece.index)
		piece.timeout = time.Now()
		piece.PieceManager()
		// Don't ask the same peer for too many pieces
		piece.pending.Wait()
	}

}

func (piece *Piece) PieceManager() {
	// TODO: Set as global?
	numberOfBlocks := Torrent.Info.PieceLength / int64(blocksize)
ManageBlocks:
	for {
		var blocks []*Block
	WaitBlocks:
		select {
		case block := <-piece.chanBlocks:
			blocks[int(block.offset)/blocksize] = block
			if len(blocks) == int(numberOfBlocks) {
				break WaitBlocks
			} else {
				continue ManageBlocks
			}
		case <-time.After(time.Second * 30):
			// Shit got fucked up
			return
		default:
			continue ManageBlocks
		}

	}
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
