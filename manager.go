package main

func Run() {
	//completion.Sync.Add(len(Pieces))
	debugger.Println("Running with This many peers", len(Peers))
	go FillQueue()    // Fill PieceQueu
	go ConnectPeers() // Fill PeerQueue

	for {
		piece := <-PieceQueue
		// Check if next peer has piece,
		// otherwise add it back to the queue and
		peer := <-PeerQueue
		go piece.AskAPeer(peer)
		PeerQueue <- peer

	}
}

func (piece *Piece) AskAPeer(peer *Peer) {

}

/* VERSION ONE */
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

func (p *Peer) AskForData() {
	p.ListenWg.Add(1)
	go p.ListenToPeer()
	p.ListenWg.Wait()
	p.sendStatusMessage(InterestedMsg)
	p.ChokeWg.Wait()
	for {
		if !p.Alive {
			debugger.Printf("Peer %s is nolonger alive: %s", p.Id, p.Ip)
			break
		}
		// TODO: if peer.has piece
		// if not, put piece back into Queue
		piece := <-PieceQueue
		//debugger.Println(piece.hash, piece.index)
		p.requestPiece(piece.index)
		//piece.askForPiece(p)
		piece.status = 1
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
