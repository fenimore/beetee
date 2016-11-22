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
	logger.Println("Listening on Server")
	logger.Println(sv.listener.Addr())
	for {
		// Wait for a connection.
		conn, err := sv.listener.Accept()
		if err != nil {
			debugger.Println(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			handshake := make([]byte, 68)
			_, err := io.ReadFull(c, handshake)
			if err != nil {
				debugger.Printf("Error Listening: %s", err)
			}
			debugger.Println(handshake)
			c.Close()
		}(conn)
	}
}
