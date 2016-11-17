package main

// Flood is when the client run
func Flood() {
	completionSync.Add(len(Pieces))
	debugger.Println("This many peers", len(Peers))
	go FillQueue()
	go ConnectPeers()

	for {

		peer := <-PeerQueue
		go peer.AskForData()
	}
}

func (peer *Peer) AskForData() {
	peer.ListenWg.Add(1)
	go peer.ListenToPeer()
	peer.ListenWg.Wait()
	peer.sendStatusMessage(InterestedMsg)
	peer.ChokeWg.Wait()
	for {
		if !peer.Alive {
			debugger.Printf("Peer %s is nolonger alive: %s", peer.Id, peer.Ip)
			break
		}
		// TODO: if peer.has piece
		// if not, put piece back into Queue
		piece := <-PieceQueue
		//debugger.Println(piece.hash, piece.index)
		//p.requestPiece(piece.index)
		piece.Pending.Add(1)
		piece.AskForPiece(peer)
		piece.Pending.Wait()
	}
}

func ConnectPeers() {
	for _, peer := range Peers {
		go func(p *Peer) {
			err := p.ConnectToPeer()
			if err == nil {
				PeerQueue <- p
			} else {
				debugger.Printf("Error Connecting to  %s: %s", p.Ip, err)
			}
		}(peer)
	}
}

// FillQueue fills the channel for asking for pieces.
func FillQueue() {
	order := DecidePieceOrder()
	for _, val := range order {
		PieceQueue <- Pieces[val]
	}
	queueSync.Done()
}

// TODO:
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
