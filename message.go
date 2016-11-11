package main

var Protocol = [19]byte{'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l'}

const (
	Choke = iota
	Unchoke
	Interested
	NotInterested
	Have
	BitField
	Request
	Piece
	Cancel
	Port
)
