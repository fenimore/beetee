package main

import "testing"
import "bytes"
import "os"
import "github.com/anacrolix/torrent/bencode"
import "encoding/binary"
import "fmt"

func TestMain(m *testing.M) {
	PeerId = GenPeerId()
	Torrent, _ = ParseTorrent("torrents/tom.torrent")
	os.Exit(m.Run())
}

func TestPeerIdSize(t *testing.T) {
	peerid := GenPeerId()

	if len(peerid) != 20 {
		t.Error("Peer Id should be 20 bytes")
	}
}

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

	if len(Torrent.Info.Pieces)%20 != 0 {
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
	if !bytes.Equal(hs[48:], PeerId[:]) {
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
	if !peers[0].choked {
		t.Error("Peer Should be choked")
	}

	numPieces := len(Torrent.Info.Pieces) / 20
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

// Message Test
func TestRequestMessage(t *testing.T) {
	msg := RequestMessage(24, blocksize*3)
	fmt.Println(msg)
	if len(msg[8:]) < 1 {
		t.Error("Block is empty?")
	}
	index := binary.BigEndian.Uint32(msg[5:9])
	if index != 24 {
		t.Error("Wrong index")
	}
	begin := binary.BigEndian.Uint32(msg[9:13])
	if int(begin)/blocksize != 3 {
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
	length := binary.BigEndian.Uint32(msg[0:4])
	if length != 1 {
		t.Error("Wrong length prefix")
	}
	id := msg[4]
	if id != InterestedMsg {
		t.Error("Wrong Status")
	}
}

func TestDecodePieceMessage(t *testing.T) {
	msg := PieceMessage(2, blocksize*2, []byte("I am the payload"))
	fmt.Println(msg)
	if int(msg[4]) != RequestMsg {
		fmt.Println(msg[4])
		t.Error("Request Message ID")
	}
	b := DecodePieceMessage(msg)
	if b.index != 2004 {
		fmt.Println(b.index)
		t.Error("Piece index for block no good")
	}
	if int(b.offset) != 2*blocksize {
		fmt.Println(2*blocksize, blocksize)
		t.Error("Block offset not good")
	}
}

func ExampleStatusMessage() {
	msg := StatusMessage(UnchokeMsg)
	fmt.Println(msg)

	// Output:
	// [0 0 0 1 1]
}
