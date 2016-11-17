package main

import "bufio"
import "errors"
import "io"
import "bytes"
import "encoding/binary"
import "crypto/sha1"
import "time"

// 19 bytes

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

// ShakeHands asks another client to accept your connection.
func (p *Peer) ShakeHands() error {
	///<pstrlen><pstr><reserved><info_hash><peer_id>
	// 68 bytes long.
	var n int
	var err error
	writer := bufio.NewWriter(p.Conn)
	// Handshake message:
	pstrlen := byte(19) // or len(pstr)
	pstr := []byte{'B', 'i', 't', 'T', 'o', 'r',
		'r', 'e', 'n', 't', ' ', 'p', 'r',
		'o', 't', 'o', 'c', 'o', 'l'}
	reserved := make([]byte, 8)
	info := Torrent.InfoHash[:]
	id := peerId[:] // my peerId
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
	err = writer.Flush()
	if err != nil {
		return err
	}
	// receive confirmation

	// The response handshake
	shake := make([]byte, 68)
	n, err = io.ReadFull(p.Conn, shake)
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
	p.Id = string(shake[48:68])
	p.Shaken = true
	return nil
}

// decodeMessage is the overall decoder, send payloads here.
// TODO: pass in a channel, and constantly be decoding for the peer
// using the switch statement
func (p *Peer) decodeMessage(payload []byte) {
	// first byte is msg type
	if len(payload) < 1 {
		p.sendStatusMessage(-1)
		return
	}
	msg := payload[1:]
	//logger.Println("recieved a Message:", payload[0])
	switch payload[0] {
	case ChokeMsg:
		p.Choked = true
		//p.ChokeWg.Done() // Does this go into the Negative nUmbers?
		p.ChokeWg.Add(1)
		logger.Println("Choked", msg)
	case UnchokeMsg:
		if p.Choked {
			p.ChokeWg.Done()
			p.Choked = false
		}
		logger.Println("UnChoke", p.Id)
	case InterestedMsg:
		p.Interested = true
		logger.Println("Interested", msg)
	case NotInterestedMsg:
		p.Interested = false
		logger.Println("NotInterested", msg)
	case HaveMsg:
		logger.Println("Have", msg)
		//p.has[binary.BigEndian.Uint32(msg)] = true
	case BitFieldMsg:
		logger.Println("Bitfield", p.Id) //, msg)
		//TODO: Parse Bitfield
		// debugger.Println(len(msg))
		// Bitfield comes right after handshake
		err := p.sendStatusMessage(InterestedMsg)
		if err != nil {
			debugger.Println("Status Error: ", err)
		}
	case RequestMsg:
		logger.Println("Request", msg)
	case BlockMsg:
		//logger.Println("Piece Message Received")
		p.decodeBlockMessage(msg)
	case CancelMsg:
		logger.Println("Cancel", msg)
	case PortMsg:
		logger.Println("Port", msg)

	}
}

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

/* Block and Piece Messages */
func (p *Peer) decodeBlockMessage(msg []byte) {
	index := binary.BigEndian.Uint32(msg[:4])
	begin := binary.BigEndian.Uint32(msg[4:8])
	// Blocks...
	block := new(Block)
	block.data = msg[8:]
	block.offset = int(begin)

	// Send to channel
	if len(block.data) < 1 {
		return
	}
	Pieces[index].chanBlocks <- block
}

// DEPRECATED
// checkPieceCompletion this is the loop which
// is constantly checking whether a piece has been
// downloaded.
// Should call one time for every piece
func (p *Piece) checkPieceCompletion() {
BlockLoop:
	for {
		b := <-p.chanBlocks
		p.blocks[b.offset/BLOCKSIZE] = b
		for _, val := range p.blocks {
			if val == nil {
				continue BlockLoop
			}
		}
		break BlockLoop
	}
	var buffer bytes.Buffer
	for _, block := range p.blocks {
		buffer.Write(block.data)
	}
	if p.hash == sha1.Sum(buffer.Bytes()) {
		p.data = buffer.Bytes()
		p.have = true
		//completionSync.Done()
		p.status = Full
		p.Pending.Done()
		logger.Printf("Piece at %d is downloaded", p.index)
		return
	}
	p.status = Empty
	p.Pending.Done()

	// TODO: This part is not making much sense
	debugger.Println(p.hash, sha1.Sum(buffer.Bytes()))
	logger.Println("Failure to sha1 hash")
	// NOTE: Call again method if fail... ?
	go p.checkPieceCompletion() // In goroutine?
}

// sendStatusMessage sends the status message to peer.
// If sent -1 then a Keep alive message is sent.
func (p *Peer) sendStatusMessage(msg int) error {
	logger.Printf("Sending Status Message: %d to %s", msg, p.Id)
	var err error
	buf := make([]byte, 4)
	writer := bufio.NewWriter(p.Conn)
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
	writer := bufio.NewWriter(p.Conn)
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, 13)
	id := byte(RequestMsg)
	// payload
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, idx)
	begin := make([]byte, 4)
	binary.BigEndian.PutUint32(begin, uint32(offset))
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(BLOCKSIZE))
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

// requestBlock takes in the block index, not the offset
func (p *Peer) requestBlock(piece, block int) {
	logger.Printf("Requesting piece %d at offset: %d", piece, block)
	offset := block * BLOCKSIZE
	err := p.sendRequestMessage(uint32(piece), offset)
	if err != nil {
		debugger.Println("Problem Requesting BLock")
	}
}

func (p *Peer) requestPiece(piece int) {
	logger.Printf("Requesting piece %d from peer %s", piece, p.Id)
	for offset := 0; offset < Torrent.Info.BlocksPerPiece; offset++ {
		err := p.sendRequestMessage(uint32(piece), offset*BLOCKSIZE)
		if err != nil {
			debugger.Println("Error Requesting", err)
		}
	}
	// Make sure to check for it's completion
	Pieces[piece].Pending.Add(1)
	Pieces[piece].status = Pending
	go Pieces[piece].checkPieceCompletion()
}

// AskForPiece takes in a peer and asks for that piece from them.
func (p *Piece) AskForPiece(peer *Peer) {
	logger.Printf("Asking %s for piece %d", peer.Id, p.index)
	//p.Lock() // Lock locks for writing, not reading
	// Calculate how many blocks, there will be.
	p.Lock() // Write lock // NOTE: Are these necessary?
	p.timeCalled = time.Now()
	p.status = Pending
	p.Unlock()
	for offset := 0; offset < Torrent.Info.BlocksPerPiece; offset++ {
		err := peer.sendRequestMessage(uint32(p.index), offset*BLOCKSIZE)
		if err != nil {
			debugger.Println("Error Requesting", err)
		}
	}
	for len(p.chanBlocks) < cap(p.chanBlocks) {
		// TODO: Set  Timeout
		// Wait
		if 30*time.Second < time.Since(p.timeCalled) {
			debugger.Println("One Minute Passed!", p.index)
			p.Lock()
			p.status = Empty
			p.Unlock()
			p.Pending.Done()
			PieceQueue <- p
			return
		}
	}
	for {
		b := <-p.chanBlocks
		p.blocks[b.offset/BLOCKSIZE] = b
		if len(p.chanBlocks) < 1 {
			break
		}
	}
	var buffer bytes.Buffer
	for _, block := range p.blocks {
		buffer.Write(block.data)
	}
	if p.hash == sha1.Sum(buffer.Bytes()) {
		p.data = buffer.Bytes()
		p.have = true
		p.Lock()
		p.status = Full
		p.Unlock()
		p.Pending.Done()
		logger.Printf("Piece at %d is downloaded", p.index)
		return
	}
	//p.Unlock()
	p.Lock()
	p.status = Empty
	p.Unlock()
	p.Pending.Done()

	// TODO: This part is not making much sense
	debugger.Println(p.hash, sha1.Sum(buffer.Bytes()))
	logger.Println("Failure to sha1 hash")
}
