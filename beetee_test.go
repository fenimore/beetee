package main

import "testing"
import "bytes"
import "os"
import "github.com/anacrolix/torrent/bencode"
import "encoding/binary"
import "fmt"
import "net"
import "io/ioutil"
import "log"

func TestMain(m *testing.M) {
	//PeerId = GenPeerId()
	d = new(Download)
	logger = log.New(os.Stdout, "Log:", log.Ltime)
	debugger = log.New(os.Stdout, "Log:", log.Ltime)
	d.Torrent, _ = ParseTorrent("torrents/tom.torrent")
	os.Exit(m.Run())
}

func TestPeerIdSize(t *testing.T) {
	peerid := GenPeerId()

	if len(peerid) != 20 {
		t.Error("Peer Id should be 20 bytes")
	}
}

// Parsing Tests
func TestPieceLen(t *testing.T) {
	tr := TrackerResponse{}
	file, err := os.Open("data/announce")
	if err != nil {
		t.Error("Couldn't open announce file")
	}
	defer file.Close()

	dec := bencode.NewDecoder(file)
	err = dec.Decode(&tr)
	if err != nil {
		debugger.Println("Unable to Decode Response")
	}

	if len(d.Torrent.Info.Pieces)%20 != 0 {
		t.Error("Pieces should be mod 20")
	}

}

func TestTorrentParse(t *testing.T) {
	_, err := ParseTorrent("torrents/tom.torrent")
	if err != nil {
		t.Error("Unable to Parse")
	}
}

// Handshake Tests
func TestHandShakeInfoHash(t *testing.T) {
	info, _ := ParseTorrent("torrents/tom.torrent")
	hs := HandShake(info)
	if !bytes.Equal(hs[28:48], info.InfoHash[:]) {
		t.Error("Incorrect infohash")
	}
}

func TestHandShakePeerId(t *testing.T) {
	info, _ := ParseTorrent("torrents/tom.torrent")
	hs := HandShake(info)
	if !bytes.Equal(hs[48:], info.PeerId[:]) {
		t.Error("Incorrect peerid")
	}
}

// Peer tests
func TestPeerParse(t *testing.T) {
	tr := TrackerResponse{}
	file, err := os.Open("data/announce")
	if err != nil {
		t.Error("Couldn't open announce file")
	}
	defer file.Close()

	dec := bencode.NewDecoder(file)
	err = dec.Decode(&tr)
	if err != nil {
		debugger.Println("Unable to Decode Response")
	}

	peers := ParsePeers(tr)
	if len(peers) != 2 {
		t.Error("Not enough Peers")
	}
	if !peers[0].choke {
		t.Error("Peer Should be choked")
	}

	numPieces := len(d.Torrent.Info.Pieces) / 20
	var expectedBitfieldSize int
	if numPieces%8 == 0 {
		expectedBitfieldSize = numPieces / 8
	} else {
		expectedBitfieldSize = numPieces/8 + 1
	}

	if len(peers[0].bitfield) != expectedBitfieldSize {
		t.Error("Bitfield is not that right size")
	}
}

// Message Tests
func TestRequestMessage(t *testing.T) {
	msg := RequestMessage(24, BLOCKSIZE*3)
	if len(msg[8:]) < 1 {
		t.Error("Block is empty?")
	}
	index := binary.BigEndian.Uint32(msg[5:9])
	if index != 24 {
		t.Error("Wrong index")
	}
	begin := binary.BigEndian.Uint32(msg[9:13])
	if int(begin)/BLOCKSIZE != 3 {
		t.Error("Wrong offset")
	}
}

func TestStatusMessage(t *testing.T) {
	// looks like thi
	//[0 0 0 1 2]
	msg := StatusMessage(InterestedMsg)
	if len(msg) != 5 {
		t.Error("Msg is too short")
	}
	length := binary.BigEndian.Uint32(msg[:4])
	if length != 1 {
		t.Error("Wrong length prefix")
	}
	id := msg[4]
	if id != InterestedMsg {
		t.Error("Wrong Status")
	}
}

func TestPieceMessage(t *testing.T) {
	// <len=0009+X><id=7><index><begin><block>
	msg := PieceMessage(2, BLOCKSIZE*2, []byte("I am the payload"))

	if int(msg[4]) != BlockMsg {
		fmt.Println(msg[4])
		t.Error("Request Message ID")
	}

	length := binary.BigEndian.Uint32(msg[:4])
	if length != uint32(len(msg[4:])) {
		t.Error("Wrong length prefix")
	}

	index := binary.BigEndian.Uint32(msg[5:9])
	if int(index) != 2 {
		t.Error("Wrong index")
	}

	offset := binary.BigEndian.Uint32(msg[9:13])
	if int(offset) != BLOCKSIZE*2 {
		fmt.Println(offset, BLOCKSIZE*2)
		t.Error("Wrong offset")
	}
}

func TestDecodePieceMessage(t *testing.T) {
	msg := PieceMessage(2, BLOCKSIZE*2, []byte("I am the payload"))

	b := DecodePieceMessage(msg[4:])
	if b.index != 2 {
		t.Error("Piece index for block no good")
	}
	if int(b.offset) != 2*BLOCKSIZE {
		fmt.Println(2*BLOCKSIZE, BLOCKSIZE)
		t.Error("Block offset not good")
	}
	if string(b.data) != "I am the payload" {
		fmt.Println("Wrong payload")
	}
}

func TestConnect(t *testing.T) {
	peer := Peer{
		addr: ":6882",
	}
	msg := StatusMessage(InterestedMsg)

	l, err := net.Listen("tcp", ":6882")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		err := peer.Connect()
		if err != nil {
			t.Error(err)
		}
		defer peer.conn.Close()
		peer.conn.Write(msg)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf, msg) {
			t.Error("Message sent wasn't received")
		}

		return // Done
	}
}

func ExampleStatusMessage() {
	msg := StatusMessage(UnchokeMsg)
	fmt.Println(msg)

	// Output:
	// [0 0 0 1 1]
}

func ExamplePieceMessage() {
	// <len=0009+X><id=7><index><begin><block>
	msg := PieceMessage(2, BLOCKSIZE*2, []byte("I am the payload"))
	fmt.Println(string(msg[13:]))

	//output:
	// I am the payload
}
