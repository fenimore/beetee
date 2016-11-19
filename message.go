package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
)

const (
	ChokeMsg = iota
	UnchokeMsg
	InterestedMsg
	NotInterestedMsg
	HaveMsg
	BitFieldMsg
	RequestMsg
	BlockMsg // rather than PieceMsg
	CancelMsg
	PortMsg
)

/*###################################################
Recieving Messages
######################################################*/

func (p *Peer) DecodeMessages(recv <-chan []byte) {
	for {
		var payload []byte
		select {
		case <-p.stopping:
			debugger.Printf("Peer %s is closing", p.id)
			p.conn.Close()
			p.alive = false
			return
		case payload = <-recv:
			//debugger.Println("Received")
		}
		//payload := <-recv
		if len(payload) < 1 {
			continue
		}
		msg := payload[1:]
		switch payload[0] {
		case ChokeMsg:
			if !p.choking {
				p.choking = true
				p.choke.Add(1)
			}
			logger.Printf("Recv: %s sends choke", p.id)
		case UnchokeMsg:
			if p.choking {
				p.choking = false
				p.choke.Done()
			}
			logger.Printf("Recv: %s sends unchoke", p.id)
		case InterestedMsg:
			p.interested = true
			logger.Printf("Recv: %s sends interested", p.id)
		case NotInterestedMsg:
			p.interested = false
			logger.Printf("Recv: %s sends uninterested", p.id)
		case HaveMsg:
			// TODO: Set bitfield
			logger.Printf("Recv: %s sends have %b", p.id, msg)
		case BitFieldMsg:
			//TODO: Parse Bitfield
			logger.Printf("Recv: %s sends bitfield %v",
				p.id, msg)
		case RequestMsg:
			logger.Printf("Recv: %s sends request %s", p.id, msg)
		case BlockMsg: // Officially "Piece" message
			// TODO: Remove this message, as they are toomuch
			logger.Printf("Recv: %s sends block", p.id)
			p.decodePieceMessage(msg)
		case CancelMsg:
			logger.Printf("Recv: %s sends cancel %s", p.id, msg)
		case PortMsg:
			logger.Printf("Recv: %s sends port %s", p.id, msg)
		default:
			continue

		}
	}
}

func (p *Peer) decodePieceMessage(msg []byte) {
	if len(msg[8:]) < 1 {
		return
	}
	index := binary.BigEndian.Uint32(msg[:4])
	begin := binary.BigEndian.Uint32(msg[4:8])
	data := msg[8:]
	// Blocks...
	block := &Block{index: index, offset: begin, data: data}
	Pieces[index].chanBlocks <- block

	if len(Pieces[index].chanBlocks) == cap(Pieces[index].chanBlocks) {
		Pieces[index].writeBlocks()
	}
}

func (p *Piece) writeBlocks() {
	p.pending.Done() // not waiting on any more blocks
	if len(p.chanBlocks) < cap(p.chanBlocks) {
		logger.Printf("The block channel for %d is not full", p.index)
		return
	}
	for {
		b := <-p.chanBlocks // NOTE: b for block
		// Copy block data to p
		// NOTE: If this doesn't work,
		// Go back to old manner of using a indexed in
		// block array
		copy(p.data[int(b.offset):int(b.offset)+blocksize],
			b.data)
		if len(p.chanBlocks) < 1 {
			break
		}
	}
	if p.hash != sha1.Sum(p.data) {
		p.data = nil
		p.data = make([]byte, p.size)
		logger.Printf("Unable to Write Blocks to Piece %d",
			p.index)
		return
	}
	p.verified = true
	logger.Printf("Piece at %d is successfully written", p.index)
	ioChan <- p
}

// 19 bytes
func (p *Peer) decodeHaveMessage(msg []byte) {
}

func (p *Peer) decodeBitfieldMessage(msg []byte) {
}

func (p *Peer) decodeRequestMessage(msg []byte) {
}

func (p *Peer) decodeCancelMessage(msg []byte) {
}

func (p *Peer) decodePortMessage(msg []byte) {
}

/*###################################################
Sending Messages
######################################################*/

// sendHandShake asks another client to accept your connection.
func (p *Peer) sendHandShake() error {
	///<pstrlen><pstr><reserved><info_hash><peer_id>
	// 68 bytes long.
	var n int
	var err error
	writer := bufio.NewWriter(p.conn)
	// Handshake message:
	pstrlen := byte(19) // or len(pstr)
	pstr := []byte{'B', 'i', 't', 'T', 'o', 'r',
		'r', 'e', 'n', 't', ' ', 'p', 'r',
		'o', 't', 'o', 'c', 'o', 'l'}
	reserved := make([]byte, 8)
	info := Torrent.InfoHash[:]
	id := PeerId[:] // my peerId NOTE: Global
	// Send handshake message
	err = writer.WriteByte(pstrlen)
	if err != nil {
		return err
	}
	n, err = writer.Write(pstr)
	if err != nil || n != len(pstr) {
		return err
	}
	n, err = writer.Write(reserved)
	if err != nil || n != len(reserved) {
		return err
	}
	n, err = writer.Write(info)
	if err != nil || n != len(info) {
		return err
	}
	n, err = writer.Write(id)
	if err != nil || n != len(id) {
		return err
	}
	err = writer.Flush() // TODO: Do I need to Flush?
	if err != nil {
		return err
	}
	// receive confirmation

	// The response handshake
	shake := make([]byte, 68)
	// TODO: Does this block?
	// TODO: Set deadline?
	n, err = io.ReadFull(p.conn, shake)
	if err != nil {
		return err
	}
	// TODO: Check for Length
	if !bytes.Equal(shake[1:20], pstr) {
		return errors.New("Protocol does not match")
	}
	if !bytes.Equal(shake[28:48], info) {
		return errors.New("InfoHash Does not match")
	}
	p.id = string(shake[48:68])
	return nil
}

// sendStatusMessage sends the status message to peer.
// If sent -1 then a Keep alive message is sent.
func (p *Peer) sendStatusMessage(msg int) error {
	logger.Printf("Sending Status Message: %d to %s", msg, p.id)
	var err error
	buf := make([]byte, 4)
	writer := bufio.NewWriter(p.conn)
	if msg == -1 { // keep alive, do nothing TODO: add ot iota
		binary.BigEndian.PutUint32(buf, 0)
	} else {
		binary.BigEndian.PutUint32(buf, 1)
	}
	writer.Write(buf)
	if err != nil {
		return err
	}
	switch msg { //<len=0001><id=0>
	case ChokeMsg:
		err = writer.WriteByte((uint8)(0))
	case UnchokeMsg:
		err = writer.WriteByte((uint8)(1))
	case InterestedMsg:
		err = writer.WriteByte(byte(2))
	case NotInterestedMsg:
		err = writer.WriteByte((uint8)(3))
	}
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

// sendRequestMessage pass in the index of the piece your looking for,
// and the offset of the piece (it's offset index * BLOCKSIZE
func (p *Peer) sendRequestMessage(idx uint32, offset int) error {
	//4-byte message length,1-byte message ID, and payload:
	// <len=0013><id=6><index><begin><length>
	// NOTE: being offset the offset by byte:
	// that is  0, 16K, 13K, etc
	var err error
	writer := bufio.NewWriter(p.conn)
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, 13)
	id := byte(RequestMsg)
	// payload
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, idx)
	begin := make([]byte, 4)
	binary.BigEndian.PutUint32(begin, uint32(offset))
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(blocksize))
	_, err = writer.Write(len)
	if err != nil {
		return err
	}
	err = writer.WriteByte(id)
	if err != nil {
		return err
	}
	_, err = writer.Write(index)
	if err != nil {
		return err
	}
	_, err = writer.Write(begin)
	if err != nil {
		return err
	}
	_, err = writer.Write(length)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

// FOR TESTING NOTE
func (p *Peer) requestAllPieces() {
	total := len(Pieces)
	//completionSync.Add(total - 1)
	debugger.Printf("Requesting all %d pieces", total)
	for i := 0; i < total; i++ {
		p.requestPiece(i)
	}
}

func (p *Peer) requestPiece(piece int) {
	logger.Printf("Requesting piece %d from peer %s", piece, p.id)
	blocksPerPiece := int(Torrent.Info.PieceLength) / blocksize
	for offset := 0; offset < blocksPerPiece; offset++ {
		err := p.sendRequestMessage(uint32(piece), offset*blocksize)
		if err != nil {
			debugger.Println("Error Requesting", err)
		}
	}
}
