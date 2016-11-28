package main

import (
	"fmt"
	"io"
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
	logger.Println("Launching New Server")
	port := 6882
	server := &Server{
		connecting: make(chan *Peer),
		stopping:   make(chan bool),
		port:       port,
		laddr:      fmt.Sprintf(":%d", port),
	}
	var err error
	server.listener, err = net.Listen("tcp4", server.laddr)
	if err != nil {
		debugger.Fatalf("FATAL, Server Fails: %s", err)
	}

	return server
}

// Listen will listen for connections and say HI!
func (sv *Server) Listen() {
	// TODO: Make Channel of accepted peers
	logger.Println("Listening on Server")
	logger.Println(sv.listener.Addr())
	for {
		// Wait for a connection.
		conn, err := sv.listener.Accept()
		if err != nil {
			debugger.Println(err)
		}
		p := &Peer{
			conn:     conn,
			port:     90,
			addr:     conn.LocalAddr().String(),
			choked:   true,
			choking:  make(chan bool),
			stopping: make(chan bool),
			bitfield: make([]bool, len(Pieces)),
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(peer *Peer) {
			// Check Handshake
			handshake := make([]byte, 68)
			_, err := io.ReadFull(peer.conn, handshake)
			if err != nil {
				debugger.Printf("Error reading handshake: %s", err)
			}
			err = peer.decodeHandShake(handshake)
			if err != nil {
				debugger.Printf("Error Handshaking: %s", err)
			}
			// TODO: Add Peer to Accepted Channels
			// TODO: Send bitfield
			debugger.Println("New Peer %s", peer.id)
			peer.ListenPeer()
		}(p)
	}
}
