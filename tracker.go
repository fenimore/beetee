package main

import "github.com/anacrolix/torrent/bencode"
import "strconv"
import "net"

//import "bytes"

import "net/http"

//import "strings"

//import "io/ioutil"

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
	TrackerId     string        `bencode:"tracker id"`
	Complete      int32         `bencode:"complete"`
	Incomplete    int32         `bencode:"incomplete"`
	Peers         bencode.Bytes `bencode:"peers"`
	PeerList      []*Peer
}

// type TrackerResponse struct {
//	Failure     string        `bencode:"failure reason"`
//	Interval    int64         `bencode:"interval"`
//	IntervalMin int64         `bencode:"min interval"`
//	TrackerId   string        `bencode:"tracker id"`
//	Complete    int64         `bencode:"complete"`
//	Incomplete  int64         `bencode:"incomplete"`
//	Peers       bencode.Bytes `bencode:"peers"`
//	//Peers Peer `bencode:"peers"`
//	PeerDict []Peer
// }

type TrackerResponseDict struct {
	Failure  string `bencode:"failure reason"`
	Interval int64  `bencode:"interval"`
}

// GetTrackerResponse TODO: pass in TrackerRequest instead
func GetTrackerResponse(m TorrentMeta) (TrackerResponse, error) { //(map[string]interface{}, error) {
	var response = TrackerResponse{} //make(map[string]interface{})

	// TODO: Use scrape conventions
	//url := strings.Replace(m.Announce, "announce", "scrape", 1)
	//request := url + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + GenPeerId() +
	request := m.Announce + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + UrlEncode(peerId) +
		"&uploaded=0" +
		"&downloaded=0" +
		"&left=" + strconv.Itoa(int(m.Info.Length)) +
		"&port=51413" +
		"&key=60502143" + "&numwant=80&compact=1&supportcrypto=1" +
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
	//fmt.Println(string(resp.Body))
	if err != nil {
		debugger.Println("Unable to Decode Response")
		return response, err
	}

	var start int
	for idx, val := range response.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}

	p := response.Peers[start:]
	for i := 0; i < len(p); i = i + 6 {
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{Ip: ip.String(), Port: port}
		peer.meta = &m
		peer.Choked = true
		response.PeerList = append(response.PeerList, &peer)
	}

	return response, err
}

func GenPeerId() [20]byte {
	// TODO: Make random
	b := [20]byte{'-', 'T', 'R', '3', '4', '4', '0', '-',
		'9', '1', 'a', '2', '4', 'W', '5', '7', '7', '4', '6', '1'}
	return b
}
