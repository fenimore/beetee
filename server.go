package main

import (
	"fmt"
	"net"
)

type Server struct {
	port     int
	listener net.Listener
	laddr    string
	// Channels
	connecting chan *Peer
	stopping   chan bool
	// Stats
	peers []*Peer
}

func NewServer() *Server {
	port := 6882
	server := &Server{
		connecting: make(chan *Peer),
		stopping:   make(chan bool),
		port:       port,
		laddr:      fmt.Sprintf(":%d", port),
	}
	var err error
	server.listener, err = net.Listen("tcp", server.laddr)
	if err != nil {
		debugger.Println("FATAL, Server Fails")
	}

}
