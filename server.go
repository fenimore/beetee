package main

import (
	"net"
)

type Server struct {
	Port     int
	Listener *net.TCPListener
	// Channels
	connecting chan *Peer
	stopping   chan bool
	// Stats
	peers []*Peer
}

func NewServer() *Server {
	server := &Server{
		connecting: make(chan *Peer),
		stopping:   make(chan bool),
	}

}
