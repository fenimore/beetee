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
	if len(msg[8:]) < 1 {
		t.Error("Block is empty?")
	}
	index := binary.BigEndian.Uint32(msg[5:9])
	if index != 24 {
		fmt.Println(index, msg[5:9])
		t.Error("Wrong index")
	}
	begin := binary.BigEndian.Uint32(msg[9:13])
	if int(begin)/blocksize != 3 {
		fmt.Println(begin, msg[9:13])
		t.Error("Wrong offset")
	}
}
