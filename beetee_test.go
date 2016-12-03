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
import "io"

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

func TestTorrentParseMultiple(t *testing.T) {
	meta, err := ParseTorrent("torrents/tails.torrent")
	if err != nil {
		t.Error(err)
	}
	if meta.Info.SingleFile {
		t.Error("Multiple files")
	}
	if len(meta.Info.Pieces)/20 != len(d.Pieces) {
		t.Error("Error parsing pieces")
	}
	lastSize := meta.Info.Length % meta.Info.PieceLength
	if d.Pieces[len(d.Pieces)-1].size != lastSize {
		t.Error("Last piece size is off")
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

func TestDecodeRequestMessage(t *testing.T) {
	msg := RequestMessage(24, BLOCKSIZE*3)
	idx, offset, _ := DecodeRequestMessage(msg[4:])

	if idx != 24 {
		t.Error("Wrong index")
	}
	if offset/BLOCKSIZE != 3 {
		t.Error("Wrong offset")
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

func TestHandShake(t *testing.T) {
	peer := Peer{
		addr: ":6883",
		info: d.Torrent,
	}

	l, err := net.Listen("tcp", ":6883") // the listening peer?
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
		err = peer.HandShake()
		if err != nil {
			t.Error(err)
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		shake := make([]byte, 68)
		_, err = io.ReadFull(conn, shake)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(shake[1:20], pstr) {
			t.Error("Protocol does not match")
		}
		if !bytes.Equal(shake[28:48], d.Torrent.InfoHash[:]) {
			t.Error("InfoHash Does not match")
		}
		hs := HandShake(d.Torrent)
		conn.Write(hs[:])

		return // Done
	}
}

func TestServeHandShake(t *testing.T) {
	_ = Serve(6884, make(chan bool)) // NOTE: diff port than prod
	peer := Peer{                    // This is my "Server"
		addr: ":6884",
		info: d.Torrent,
	}
	peer.Connect()
	err := peer.HandShake()
	if err != nil {
		t.Error(err)
	}

}

func TestUnchokeLeecher(t *testing.T) {
	_ = Serve(6886, make(chan bool)) // NOTE: diff port than prod
	peer := Peer{                    // This is my "Server"
		addr: ":6886",
		info: d.Torrent,
	}
	peer.Connect()
	err := peer.HandShake()
	if err != nil {
		t.Error(err)
	}

	// TODO: Set up listening for message
	// from server

	msg := StatusMessage(InterestedMsg)
	peer.conn.Write(msg)
}

// func TestIOMultipleFile(t *testing.T) {
//	firstPayload := []byte("I am Payload")

//	piece := &Piece{
//		data:  firstPayload,
//		index: 1,
//		size:  int64(len(firstPayload)),
//	}

//	files := []*TorrentFile{
//		&TorrentFile{
//			Length: 6},
//		&TorrentFile{
//			Length: 3},
//		&TorrentFile{
//			Length: 5},
//	}
//	var total int64 // length of files
//	for _, file := range files {
//		file.PreceedingTotal = total
//		total += file.Length
//	}

//	file := files[1]
//	_, data, _ := oldPieceInFile(piece, file)

//	fmt.Println(string(data))
//	if string(data) != "What" {
//		t.Error("The Space hasn't been filled")
//	}

// }

func TestIOMultipleFilesNew(t *testing.T) {
	payload := []byte("iameight")

	piece := &Piece{
		data:  payload,
		index: 1,
		size:  int64(len(payload)),
	}
	files := []*TorrentFile{
		&TorrentFile{
			Path:   []string{"Dir/", "File"},
			Length: 12},
		&TorrentFile{
			Path:   []string{"Dir/", "Info"},
			Length: 18},
		&TorrentFile{
			Path:   []string{"Dir/", "Other"},
			Length: 6},
		&TorrentFile{
			Path:   []string{"Dir/", "End"},
			Length: 10},
	}
	// Get Preceeding Total of Files
	var total int64 // length of files
	for _, file := range files {
		file.PreceedingTotal = total
		total += file.Length
	}

	result := make([]byte, 0)

	for _, file := range files {
		ok, data, offset := pieceInFile(piece, file, piece.size)
		if ok {
			//fmt.Println(len(data), offset, file.Path)
			//fmt.Println(string(data))
			fmt.Println("Offset:", offset, file.Path, string(data), len(data))
			result = append(result, data...)
		}
	}
	if !bytes.Equal(result, payload) {
		t.Error("Payload wasn't 'written' to 'file'")
	}
}

func TestIOMultipleFilesRealistic(t *testing.T) {
	payload := []byte("iameight")

	piece := &Piece{
		data:  payload,
		index: 29, // 16, // 22 and 29 are edges too // On the edge: 11, // 12
		size:  int64(len(payload)),
	}
	files := []*TorrentFile{
		&TorrentFile{
			Path:   []string{"Dir/", "File"},
			Length: 96},
		&TorrentFile{
			Path:   []string{"Dir/", "Info"},
			Length: 84},
		&TorrentFile{
			Path:   []string{"Dir/", "Other"},
			Length: 57},
		&TorrentFile{
			Path:   []string{"Dir/", "End"},
			Length: 10},
	}
	// Get Preceeding Total of Files
	var total int64 // length of files
	for _, file := range files {
		file.PreceedingTotal = total
		total += file.Length
	}

	result := make([]byte, 0)

	for _, file := range files {
		ok, data, offset := pieceInFile(piece, file, piece.size)
		if ok {
			//fmt.Println(len(data), offset, file.Path)
			//fmt.Println(string(data))
			fmt.Println("Offset:", offset, file.Path, string(data), len(data))
			result = append(result, data...)
		}
	}
	if !bytes.Equal(result, payload) {
		t.Error("Payload wasn't 'written' to 'file'")
	}
}

func TestIOSmallFirstFile(t *testing.T) {
	payload := []byte("iameight")
	piece := &Piece{
		data:  payload,
		index: 0,
		size:  int64(len(payload)),
	}
	files := []*TorrentFile{
		&TorrentFile{
			Path:   []string{"Dir/", "File"},
			Length: 2},
		&TorrentFile{
			Path:   []string{"Dir/", "Info"},
			Length: 18},
	}
	// Get Preceeding Total of Files
	var total int64 // length of files
	for _, file := range files {
		file.PreceedingTotal = total
		total += file.Length
	}

	result := make([]byte, 0)

	for _, file := range files {
		ok, data, offset := pieceInFile(piece, file, piece.size)

		if ok {
			fmt.Println("Offset:", offset, file.Path, string(data), len(data))
			result = append(result, data...)
		}
	}
	if !bytes.Equal(result, payload) {
		fmt.Println(string(result))
		t.Error("Payload wasn't 'written' to 'file'")
	}
}

// Works well when the piece is bigger than the
// the total of files
func TestIOMultipleFiles(t *testing.T) {
	//firstPayload := []byte("iameightplusplus")
	firstPayload := []byte("iameight")
	piece := &Piece{
		data:  firstPayload,
		index: 3,
		size:  int64(len(firstPayload)),
	}
	files := []*TorrentFile{
		&TorrentFile{
			Path:   []string{"Dir/", "File"},
			Length: 12},
		&TorrentFile{
			Path:   []string{"Dir/", "Info"},
			Length: 18},
		&TorrentFile{
			Path:   []string{"Dir/", "Other"},
			Length: 6},
		&TorrentFile{
			Path:   []string{"Dir/", "End"},
			Length: 5},
		&TorrentFile{
			Path:   []string{"Dir/", "Appendix"},
			Length: 5},
	}
	// Get Preceeding Total of Files
	var total int64 // length of files
	for _, file := range files {
		file.PreceedingTotal = total
		total += file.Length
	}

	result := make([]byte, 0)

	for _, file := range files {
		ok, data, offset := pieceInFile(piece, file, piece.size)

		if ok {
			fmt.Println("Offset:", offset, file.Path, string(data), len(data))
			result = append(result, data...)
		}
	}

	if !bytes.Equal(result, firstPayload) {
		fmt.Println(string(result))
		fmt.Println(len(result), total)
		t.Error("Payload wasn't 'written' to 'file'")
	}
}

func TestIOMultipleFilesBigger(t *testing.T) {
	//firstPayload := []byte("iameightplusplus")
	dummy_data := make([]byte, 131072)
	for i, _ := range dummy_data {
		dummy_data[i] = 1
	}

	pieces := make([]*Piece, 0)
	for i := 0; i < 128; i++ {
		pieces = append(pieces, &Piece{
			data:  dummy_data,
			index: i,
			size:  131072,
		})
	}

	files := []*TorrentFile{
		&TorrentFile{
			Path:   []string{"Dir/", "File"},
			Length: 804},
		&TorrentFile{
			Path:   []string{"Dir/", "Info"},
			Length: 7942424},
		&TorrentFile{
			Path:   []string{"Dir/", "Other"},
			Length: 8945491},
	}
	// Get Preceeding Total of Files
	var total int64 // length of files
	for _, file := range files {
		file.PreceedingTotal = total
		total += file.Length
	}

	result := make([]byte, 0)
	for _, piece := range pieces {
		for _, file := range files {
			ok, data, offset := pieceInFile(piece, file, piece.size)

			if ok {
				fmt.Println("Offset:", offset, file.Path, len(data))
				result = append(result, data...)
			}
		}
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
