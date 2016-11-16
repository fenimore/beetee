package main

import (
	"sync"
)

type Manager struct {
	torrent     *TorrentMeta
	peers       []*Peer
	alives      []*Peer
	pieceQ      chan int
	pieceStatus map[int]int // empty/pending/full 0, 1, 2
	left        int
	uploaded    int
	wg          sync.WaitGroup
	// TODO: Delta for keep alive
	// TODO: delta ask for more peers
	// TODO: more things..
}

func (m *Manager) Flood() {
	for _, peer := range m.peers[1:6] {
		go peer.ListenToPeer()
		m.alives = append(m.alives, peer)
	}

	for idx, _ := range m.torrent.Info.PieceList {
		m.pieceStatus[idx] = 0
	}
	go m.AskForPieces()
	for {
		m.ShiftPieceQ()
		m.wg.Wait()
	}
}

func (m *Manager) ShiftPieceQ() {
	order := m.DecidePieceOrder()
	// TODO: Check Peers for who has what
	for i := 0; i < 5; i++ {
		m.pieceQ <- order[0]
		m.wg.Add(1)
	}

}

func (m *Manager) AskForPieces() {
	for {
		idx := <-m.pieceQ
		if m.pieceStatus[idx] != Empty {
			continue
		}
		m.pieceStatus[idx] = Pending
		peer := m.FindPeerForPiece(idx)
		peer.requestPiece(idx)
	}
}

func (m *Manager) FindPeerForPiece(idx int) *Peer {
	// TODO: find in alives who has idx m.alives
	for _, peer := range m.alives {
		if peer.Alive {
			return peer
		}
	}
	return nil
}

// DecidePieceOrder should return a list of indexes
// of pieces, according to the rarest first
func (m *Manager) DecidePieceOrder() []int {
	order := make([]int, 0, len(m.torrent.Info.PieceList))
	for i := 0; i < len(m.torrent.Info.PieceList); i++ {
		if m.pieceStatus[i] == 0 {
			order = append(order, i)
		}
	}
	return order
}
