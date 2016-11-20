package main

import (
	"crypto/sha1"
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

// PieceManager must be run for every piece..
// NOTE: Should I request pieces here?
// TODO: add pending here.
// FIXME: hash duplicates?
func (piece *Piece) PieceManager() {
	// TODO: Set as global?
	numberOfBlocks := int(Torrent.Info.PieceLength) / blocksize
	var blockCount int
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
					debugger.Println("Invalid Hash of  Piece %d",
						piece.index)
					PieceQueue <- piece
					return
				}
				break
			} else if piece.index == len(Pieces)-1 {
				// Last Piece could have fewer blocks
				if piece.hash == sha1.Sum(piece.data) {
					break
				}
			} else { // NOTE Not enough block
				continue
			}
		case <-time.After(time.Second * 30):
			debugger.Println("Piece %d timeout", piece.index)
			PieceQueue <- piece
			return
		}
		//logger.Println("Begin Writing Block")
		piece.verified = true
		logger.Printf("Piece at %d is successfully written",
			piece.index)
		ioChan <- piece

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
