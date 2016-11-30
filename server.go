package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type Server struct {
	listener net.Listener
	laddr    string
}

// Serve, TODO: this is for seeding not leeching
func Serve(port int, shutdown <-chan bool) chan<- *Peer {
	leechers := make(chan *Peer)
	incoming := make(chan net.Conn)

	var err error
	server := &Server{laddr: fmt.Sprintf(":%d", port)}

	server.listener, err = net.Listen("tcp", server.laddr)
	if err != nil {
		debugger.Fatalf("Fatal, server fails: %s", err)
	}
	go func() {
		for {

			conn, err := server.listener.Accept()
			if err != nil {
				conn.Close()
			}
			incoming <- conn
			// TODO: handleError(err)
		}
	}()
	go func() {
		for {
			newConn := <-incoming
			shake := make([]byte, 68)
			_, err = io.ReadFull(newConn, shake)
			if err != nil {
				debugger.Fatal(err)
			}
			if !bytes.Equal(shake[1:20], pstr) {
				debugger.Println("Protocol does not match")
			}
			if !bytes.Equal(shake[28:48], d.Torrent.InfoHash[:]) {
				debugger.Println("InfoHash Does not match")
			}
			peer := &Peer{conn: newConn, id: string(shake[48:]),
				addr: newConn.RemoteAddr().String()}
			hs := HandShake(d.Torrent)
			peer.conn.Write(hs[:])
			debugger.Println("New Leecher", peer.id)
			leechers <- peer

		}
	}()
	return leechers // TODO: Add shutdown
}

// download (Torrent, Peers, all your stuff)...
// protected by mutex, inteface to stuff, goes through mutex.
// go NewServer(download, port)
