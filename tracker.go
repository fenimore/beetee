package main

import "github.com/anacrolix/torrent/bencode"
import "strconv"
import "net/http"
import "net"
import "fmt"
import "math/rand"
import "time"
import "errors"
import "encoding/binary"

const ( // UDP actions
	CONREQ = iota
	ANNREQ
)

const CONNID = 0x41727101980

type TrackerRequest struct {
	InfoHash,
	PeerId,
	Port,
	Uploaded,
	Downloaded,
	Left,
	Compact,
	Event,
	NumWant,
	Key string
}

// TrackerResponse takes in the BIN response, not dict
type TrackerResponse struct {
	FailureReason string        `bencode:"failure reason"`
	Interval      int32         `bencode:"interval"`
	IntervalMin   int64         `bencode:"min interval"`
	TrackerId     string        `bencode:"tracker id"`
	Complete      int32         `bencode:"complete"`
	Incomplete    int32         `bencode:"incomplete"`
	Peers         bencode.Bytes `bencode:"peers"`
}

// TODO: When tracker responds in dict instead of BIN
type TrackerResponseDict struct {
	Failure  string `bencode:"failure reason"`
	Interval int64  `bencode:"interval"`
}

func HTTPTracker(m *TorrentMeta) (TrackerResponse, error) {
	var response = TrackerResponse{}
	// TODO: Use scrape conventions
	request := m.Announce +
		"?info_hash=" +
		m.InfoHashEnc +
		"&peer_id=" +
		UrlEncode(m.PeerId) +
		"&uploaded=0" +
		"&downloaded=0" +
		"&left=" +
		strconv.Itoa(int(m.Info.Length)) +
		"&port=" + strconv.Itoa(PORT) +
		"&key=60502143" +
		"&numwant=80&compact=1&supportcrypto=1" +
		"&event=started"

	// TODO: conStruct
	logger.Println("HTTP Tracker request:", request)

	resp, err := http.Get(request)
	if err != nil {
		debugger.Println("Request Didn't Go through")
		return response, err
	}
	defer resp.Body.Close() // Body is a ReadCloser

	// Decode bencoded Response
	dec := bencode.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		debugger.Println("Unable to Decode Response")
		return response, err
	}

	return response, err
}

func UDPTracker(m *TorrentMeta) error {
	logger.Println("Tracker Uses UDP", m.Announce[6:len(m.Announce)-9])

	tranId := GenTransactionId()
	conn, err := net.Dial("udp", m.Announce[6:len(m.Announce)-9])
	// Connection
	_, err = conn.Write(UDPTrackerRequest(tranId))
	if err != nil {
		// Continue?
		return err
	}
	trackerResponse := make([]byte, 16)
	_, err = conn.Read(trackerResponse)
	if err != nil {
		// Continue?
		return err
	}
	connId, ok := UDPValidateResponse(trackerResponse, tranId)
	if !ok {
		return errors.New("No Connection Made to UDP")
	}

	// Announce
	_, err = conn.Write(UDPAnnounceClient(connId, *m))
	if err != nil {
		// Continue?
		return err
	}
	logger.Println("Connection ID valid 2 Minutes:", connId)
	// FIXME: Read remaining bytes in ParseAnnounce
	// 4096 is a big number, but 80 peers = 480 bytes
	trackerResponse = make([]byte, 512)
	_, err = conn.Read(trackerResponse)
	if err != nil {
		return err
	}

	// Parse Announce
	peers, err := UDPParseAnnounce(trackerResponse, conn)
	if err != nil {
		return err
	}
	d.Peers = peers

	return nil
}

// GetTrackerResponse TODO: pass in TrackerRequest instead
func GetTrackerResponse(m *TorrentMeta) error {
	if m.Announce[:3] == "udp" {
		return UDPTracker(m)
	}
	tr, err := HTTPTracker(m)
	if err != nil {
		return err
	}
	d.Peers = ParsePeers(tr)
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func GenPeerId() [20]byte {
	rand.Seed(time.Now().UnixNano())
	var b [20]byte
	copy(b[:8], []byte("-FL1001-"))
	for i := range b {
		if i == 12 {
			break
		}
		b[i+8] = letters[rand.Intn(len(letters))]
	}
	return b
}

// GenTransactionId creates random number for UDP tracker
func GenTransactionId() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

func UDPTrackerRequest(transactionId uint32) []byte {
	connReq := make([]byte, 16)
	binary.BigEndian.PutUint64(connReq[:8], CONNID)
	binary.BigEndian.PutUint32(connReq[8:12], CONREQ)
	binary.BigEndian.PutUint32(connReq[12:], transactionId)
	return connReq
}

func UDPValidateResponse(response []byte, transactionId uint32) (uint64, bool) {
	action := binary.BigEndian.Uint32(response[:4])
	tranId := binary.BigEndian.Uint32(response[4:8])
	connId := binary.BigEndian.Uint64(response[8:16])
	if action != 0 || tranId != transactionId {
		return connId, false
	}
	return connId, true
}

//UDPAnnounceClient constructs announce message for UDP packet
// Should I be passing TorrentMeta pointer?
func UDPAnnounceClient(connId uint64, m TorrentMeta) []byte {
	announce := make([]byte, 98)
	binary.BigEndian.PutUint64(announce[:8], connId)
	binary.BigEndian.PutUint32(announce[8:12], ANNREQ)
	binary.BigEndian.PutUint32(announce[12:16], GenTransactionId())
	copy(announce[16:36], m.InfoHash[:])
	copy(announce[36:56], m.PeerId[:])
	binary.BigEndian.PutUint64(announce[56:64], 0)    // Downloaded
	binary.BigEndian.PutUint64(announce[64:72], 0)    // Compeleted
	binary.BigEndian.PutUint64(announce[72:80], 0)    // Uploaded
	binary.BigEndian.PutUint32(announce[80:84], 0)    // Event None
	binary.BigEndian.PutUint32(announce[84:88], 0)    // IP (def=source)
	binary.BigEndian.PutUint32(announce[88:92], 0)    // key ?
	binary.BigEndian.PutUint32(announce[92:96], 80)   // num want peers
	binary.BigEndian.PutUint16(announce[96:98], PORT) // port

	return announce
}

// UDPParseAnnounce parses a announce response and reads
// the remaining bytes from connection according to how
// many peers are designated.
func UDPParseAnnounce(resp []byte, conn net.Conn) ([]*Peer, error) {
	// The first 20 bytes are mandatory
	action := binary.BigEndian.Uint32(resp[:4])
	if action != 1 {
		return nil, errors.New("Wrong UDP tracker action")
	}
	_ = binary.BigEndian.Uint32(resp[4:8])         // transaction ID
	_ = binary.BigEndian.Uint32(resp[8:12])        // interval
	amount := binary.BigEndian.Uint32(resp[12:16]) // leechers
	amount += binary.BigEndian.Uint32(resp[16:20]) // seeders
	p := resp[20:]                                 // remaining bytes are Peer addresses
	// Keep reading from conn accordingly
	// for every peer there will be:
	// 4 - IPv4
	// 2 - Port
	// 6 * amount = bytes with relevant data
	bitCap := len(d.Pieces) / 8 // for peers
	if len(d.Pieces)%8 != 0 {
		bitCap += 1
	}

	debugger.Println("Peer Amnt", amount, int(amount)*6, p)
	var peers []*Peer
	for i := 0; i < (int(amount) * 6); i = i + 6 {
		if len(p) < i+6 || i == 480 {
			break
		}
		//debugger.Println("Index", i, i+6, len(p), len(peers))
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{
			ip:       ip.String(),
			port:     port,
			addr:     fmt.Sprintf("%s:%d", ip.String(), port),
			choke:    true,
			bitfield: make([]byte, bitCap),
			bitmap:   make([]bool, len(d.Pieces)),
			info:     d.Torrent,
			retry:    4,
		}
		peers = append(peers, &peer)
	}

	debugger.Println("Peer Cnt: ", len(peers))

	return peers, nil
}

// ParsePeers parses a bencoded response from tracker.
// Only used with HTTP tracker, though code is repeated
// in UDP parsing.
func ParsePeers(r TrackerResponse) []*Peer {
	var start int
	for idx, val := range r.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}
	p := r.Peers[start:]

	// Peer bitfields are rounded up
	bitCap := len(d.Pieces) / 8
	if len(d.Pieces)%8 != 0 {
		bitCap += 1
	}

	// A peer is represented in six bytes
	// four for ip and two for port
	var peers []*Peer
	for i := 0; i < len(p); i = i + 6 {
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{
			ip:       ip.String(),
			port:     port,
			addr:     fmt.Sprintf("%s:%d", ip.String(), port),
			choke:    true,
			bitfield: make([]byte, bitCap),
			bitmap:   make([]bool, len(d.Pieces)),
			info:     d.Torrent,
			retry:    4,
		}
		peers = append(peers, &peer)
	}
	return peers
}
