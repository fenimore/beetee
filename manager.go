package main

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

func (peer *Peer) AskPeer() {
	peer.sendStatusMessage(InterestedMsg)
	//peer.se
	for {
		select {
		case <-peer.stopping:
			return
		case <-peer.choking:
			peer.choke.Wait()
		default:
			// do nothing
		}
		peer.choke.Wait() // if Choked, then Wait
		piece := <-PieceQueue
		if !peer.bitfield[piece.index] {
			PieceQueue <- piece
			continue
		}
		piece.pending.Add(1)
		peer.requestPiece(piece.index)
		piece.pending.Wait()
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
