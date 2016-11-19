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
	peer.choke.Wait()
	//peer.se
	for {
		piece := <-PieceQueue
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
