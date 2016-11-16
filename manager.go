package main

type Manager struct {
	peers   []*Peer
	pieces  []*Piece
	torrent TorrentMeta
}
