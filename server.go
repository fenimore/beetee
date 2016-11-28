package main

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
	laddr    string
}

// Serve, TODO: this is for seeding not leeching
func Serve(port int, shutdown <-chan bool) error {
	var err error
	server := &Server{laddr: fmt.Sprintf(":%d", port)}

	server.listener, err = net.Listen("tcp", server.laddr)
	if err != nil {
		debugger.Fatalf("Fatal, server fails: %s", err)
	}
	for {

		conn, err := server.listener.Accept()
		if err == nil {
			conn.Close()
			//go handleConnection(conn, shutdown)
		}
		// TODO: handleError(err)
	}
	return err
}

// download (Torrent, Peers, all your stuff)...
// protected by mutex, inteface to stuff, goes through mutex.
// go NewServer(download, port)
