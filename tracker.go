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
	logger.Println("TRACKER REQUEST:", request)

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

func UDPTracker(m *TorrentMeta) (TrackerResponse, error) {
	var response = TrackerResponse{}
	fmt.Println("Tracker Uses UDP", m.Announce[6:len(m.Announce)-9])
	//TODO: Connect udp
	conn, err := net.Dial("udp", m.Announce[6:len(m.Announce)-9])

	fmt.Println(conn, err)
	tranId := GenTransactionId()

	// Connection
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
	// connId valid for 2 minutes
	if !ok {
		// wait some time and ask again
		fmt.Println("No tracker connection made")
	}

	// Announce
	_, err = conn.Write(UDPAnnounceClient(connId, *m))
	if err != nil {
		// Continue?
		fmt.Println(err)
	}
	// TODO: set size dynamically?
	trackerResponse = make([]byte, 256)
	n, err := conn.Read(trackerResponse)
	fmt.Println(n, "Bytes read")
	if err != nil {
		// Continue?
		fmt.Println(err)
	}

	UDPParseAnnounce(trackerResponse, conn)

	return response, errors.New("UDP not implemented yet")
}

// GetTrackerResponse TODO: pass in TrackerRequest instead
func GetTrackerResponse(m *TorrentMeta) (TrackerResponse, error) {
	if m.Announce[:3] == "udp" {
		return UDPTracker(m)
	}
	return HTTPTracker(m)
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

// TODO: return interval for next announce
func UDPParseAnnounce(resp []byte, conn net.Conn) []*Peer {
	// https://github.com/naim94a/udpt/wiki/The-BitTorrent-UDP-tracker-protocol
	// The first 20 bytes are mandatory
	// 4 - action = 1
	// 4 - transaction id
	// 4 - interval
	// 4 - leechers amount
	// 4 - seeders amount

	// Keep reading from conn accordingly
	// n = leechers + seeders amounts
	// for every peer there will be:
	// 4 - IPv4
	// 2 - Port

	return nil
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
		}
		peers = append(peers, &peer)
	}
	return peers
}
