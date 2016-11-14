package main

import "bufio"
import "errors"
import "io"
import "bytes"
import "encoding/binary"
import "crypto/sha1"

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
	info := p.meta.InfoHash[:]
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

func (p *Peer) decodeMessage(payload []byte) {
	// first byte is msg type
	if len(payload) < 1 {
		return
	}
	msg := payload[1:]
	//logger.Println("recieved a Message:", payload[0])
	switch payload[0] {
	case ChokeMsg:
		logger.Println("Choked", msg)
	case UnchokeMsg:
		logger.Println("UnChoke", msg)
		p.requestAllPieces()
	case InterestedMsg:
		logger.Println("Interested", msg)
	case NotInterestedMsg:
		logger.Println("NotInterested", msg)
	case HaveMsg:
		logger.Println("Have", msg)
		p.has[binary.BigEndian.Uint32(msg)] = false
	case BitFieldMsg:
		logger.Println("Bitfield", msg)
		// Bitfield comes right after handshake
		err := p.sendStatusMessage(InterestedMsg)
		if err != nil {
			debugger.Println("Status Error: ", err)
		}
	case RequestMsg:
		logger.Println("Request", msg)
	case BlockMsg:
		logger.Println("Piece", msg[:4])
		p.decodeBlockMessage(msg)
		go func() {
			_ = p.meta.Info.WriteData()
		}()
	case CancelMsg:
		logger.Println("Payload", msg)
	case PortMsg:
		logger.Println("Port", msg)
	}

}

func (p *Peer) decodeBlockMessage(msg []byte) {
	index := binary.BigEndian.Uint32(msg[:4])
	// Begin is which 0 based offset within the piece.
	// that is, which BLOCK this is within piece
	begin := binary.BigEndian.Uint32(msg[4:8])
	//debugger.Println("Index:", int(index), begin)
	pieceList := p.meta.Info.PieceList // for readability
	// TODO: put into Block, and then check if
	// there are other blocks to get..
	// TODO: Only if NOTE block but all of piece..
	if pieceList[index].hash == sha1.Sum(msg[8:]) {
		debugger.Printf("Valid Hash at %d offset: %d\n", index, begin)
		pieceList[index].data = msg[8:]
		pieceList[index].have = true
	} else {
		//debugger.Println("Invalid Hash :( ")
	}
}

func (p *Peer) requestPiece() {
	//countBlocks()
	// For every block in a piece
	// request that piece with offset
	//
}

func (p *Piece) checkBlockCompletion() {

}

func (p *Peer) sendStatusMessage(msg int) error {
	logger.Println("Sending Status Message: ", msg)
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

// Sendrequestmessage pass in the index of the piece your looking for.
func (p *Peer) sendRequestMessage(idx uint32, offset int) error {
	// Request lenght := 16384
	// From kristen:
	//The ‘Request’ message type consists of the
	//4-byte message length,
	//1-byte message ID,
	//and a payload composed of a
	//4-byte piece index (0 based),
	//4-byte block offset within the piece (measured in bytes), and
	//4-byte block length
	// <len=0013><id=6><index><begin><length>
	//logger.Println("Sending Request Message: ", idx)
	var err error
	writer := bufio.NewWriter(p.Conn)
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, 13)
	id := byte(RequestMsg)
	// payload
	index := make([]byte, 4)
	binary.BigEndian.PutUint32(index, idx)
	begin := make([]byte, 4)
	// Change which block to request TODO:
	binary.BigEndian.PutUint32(begin, uint32(offset))
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, 16384)

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
	total := len(p.meta.Info.PieceList)
	debugger.Println("Requesting all pieces")
	for i := 0; i < total; i++ {
		err := p.sendRequestMessage(uint32(i), 0)
		if err != nil {
			debugger.Println("Error Requesting", err)
		}
	}
}
