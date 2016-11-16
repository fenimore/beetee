package main

func Flood() {
	for _, peer := range Peers[1:6] {
		go peer.ListenToPeer()
	}
	go AskForPieces()
}

func AskForPieces() {
	for {
		//idx := <-m.pieceQ
		idx := 0
		if Pieces[idx].status != Empty {
			continue
		}
		Pieces[idx].status = Pending
		peer := FindPeerForPiece(idx)
		peer.requestPiece(idx)
	}
}

func FindPeerForPiece(idx int) *Peer {
	// TODO: find in alives who has idx m.alives
	for _, peer := range Peers {
		if peer.Alive {
			return peer
		}
	}
	return nil
}

// DecidePieceOrder should return a list of indexes
// of pieces, according to the rarest first
func DecidePieceOrder() []int {
	order := make([]int, 0, len(Pieces))
	for i := 0; i < len(Pieces); i++ {
		if Pieces[i].status == 0 {
			order = append(order, i)
		}
	}
	return order
}
