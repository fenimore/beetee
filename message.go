package main

import (
	"crypto/sha1"
	"encoding/binary"
	//"time"
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
		//Pieces[index].success <- true
	}
}

func (p *Piece) writeBlocks() {
	//p.pending.Done() // not waiting on any more blocks
	if len(p.chanBlocks) < cap(p.chanBlocks) {
		logger.Printf("The block channel for %d is not full", p.index)
		return
	}
	for {
		b := <-p.chanBlocks // NOTE: b for block
		copy(p.data[int(b.offset):int(b.offset)+blocksize],
			b.data)
		if len(p.chanBlocks) < 1 {
			break
		}
	}
	if p.hash != sha1.Sum(p.data) {
		debugger.Printf("Error with piece of size %d,\n the hash is %x, and what I got is %x", p.size, p.hash, sha1.Sum(p.data))
		p.data = nil
		p.data = make([]byte, p.size)
		logger.Printf("Unable to Write Blocks to Piece %d",
			p.index)
		return
	}
	p.verified = true
	logger.Printf("Piece at %d is successfully written", p.index)
	ioChan <- p
	p.success <- true
}

// 19 bytes
func (p *Peer) decodeHaveMessage(msg []byte) {
	//index := binary.BigEndian.Uint32(msg)
	//p.bitfield[index] = true
}

// NOTE: The bitfield will be sent with padding if the size is
// not divisible by eight.
// Thank you Tulva RC bittorent client for this algorithm
// github.com/jtakkala/tulva/
func (p *Peer) decodeBitfieldMessage(bitfield []byte) {
	// // For each byte, look at the bits
	// // NOTE: that is 8 * 8
	// for i := 0; i < len(p.bitfield); i++ {
	//	for j := 0; j < 8; j++ {
	//		index := i*8 + j
	//		if index >= len(Pieces) {
	//			break // Hit padding bits
	//		}

	//		byte := bitfield[i]              // Within bytes
	//		bit := (byte >> uint32(7-j)) & 1 // some shifting
	//		//p.bitfield[index] = bit == 1     // if bit is true
	//	}
	// }
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

var pstr = []byte("BitTorrent protocol")
var pstrlen = byte(19)

//<pstrlen><pstr><reserved><info_hash><peer_id>
// 68 bytes long.
func HandShake(info *TorrentMeta) [68]byte {
	//h := make([]byte)
	var h [68]byte
	h[0] = pstrlen
	copy(h[1:20], pstr[:])
	copy(h[28:48], info.InfoHash[:])
	copy(h[48:], PeerId[:])
	return h
}

// sendStatusMessage sends the status message to peer.
// If sent -1 then a Keep alive message is sent.

// sendRequestMessage pass in the index of the piece your looking for,
// and the offset of the piece (it's offset index * BLOCKSIZE
func RequestMessage(idx uint32, offset int) []byte {
	//4-byte message length,1-byte message ID, and payload:
	// <len=0013><id=6><index><begin><length>
	msg := make([]byte, 17)
	// Message prefix
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, 13)
	id := byte(RequestMsg)
	// Payload
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, idx)
	begin := make([]byte, 4)
	binary.BigEndian.PutUint32(begin, uint32(offset))
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(blocksize))
	// Write to buffer
	copy(msg[:4], len)
	msg[4] = id
	copy(msg[5:9], index)
	copy(msg[9:13], begin)
	copy(msg[13:17], length)

	return msg
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
		_ = RequestMessage(uint32(piece), offset*blocksize)
	}
}
