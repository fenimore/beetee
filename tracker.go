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

// GetTrackerResponse TODO: pass in TrackerRequest instead
func GetTrackerResponse(m *TorrentMeta) (TrackerResponse, error) { //(map[string]interface{}, error) {
	var response = TrackerResponse{} //make(map[string]interface{})

	// TODO: Use scrape conventions
	//url := strings.Replace(m.Announce, "announce", "scrape", 1)
	//request := url + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + GenPeerId() +
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
	logger.Println("TRACKER REQUEST:", request)

	if m.Announce[:3] == "udp" {
		fmt.Println("UDP New", m.Announce[6:len(m.Announce)-9])
		//TODO: Connect udp
		conn, err := net.Dial("udp", m.Announce[6:len(m.Announce)-9])

		fmt.Println(conn, err)
		tranId := GenTransactionId()
		fmt.Println("Transaction ID: ", tranId)
		_, err = conn.Write(UDPTrackerRequest(tranId))
		if err != nil {
			// Continue?
			fmt.Println(err)
		}
		trackerResponse := make([]byte, 16)
		_, err = conn.Read(trackerResponse)
		if err != nil {
			// Continue?
			fmt.Println(err)
		}
		connId, ok := UDPValidateResponse(trackerResponse, tranId)
		if !ok {
			// wait some time
			// and ask again
			fmt.Println("No tracker connection made")
		}
		fmt.Println(connId)
		return response, errors.New("UDP not implemented yet")
	}

	resp, err := http.Get(request)
	if err != nil {
		debugger.Println("Request Didn't Go through")
		return response, err
	}
	defer resp.Body.Close() // Body is a ReadCloser

	//fmt.Println(resp.Body)
	// Decode bencoded Response
	dec := bencode.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		debugger.Println("Unable to Decode Response")
		return response, err
	}

	return response, err
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

// parsePeers is a http response gotten from
// the tracker; parse the peers byte message
// and put to global Peers slice.
func ParsePeers(r TrackerResponse) []*Peer {
	var start int
	var peers []*Peer
	for idx, val := range r.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}
	p := r.Peers[start:]
	// A peer is represented in six bytes
	// four for ip and two for port
	bitCap := len(d.Pieces) / 8
	if len(d.Pieces)%8 != 0 {
		bitCap += 1
	}

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
			//in:       make(chan []byte),
			//out:      make(chan []byte),
			//halt:     make(chan struct{}),
		}
		peers = append(peers, &peer)
	}
	return peers
}
