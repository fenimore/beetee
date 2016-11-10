package main

import "fmt"
import "github.com/anacrolix/torrent/bencode"
import "strconv"

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

type Peer struct {
	PeerId string `bencode:"peer id"`
	Ip     string `bencode:"ip"`
	Port   int64  `bencode:"port"`
}

// GetTrackerResponse TODO: pass in TrackerRequest instead
func GetTrackerResponse(m TorrentMeta) (TrackerResponse, error) { //(map[string]interface{}, error) {
	var response = TrackerResponse{} //make(map[string]interface{})

	// TODO: Use scrape conventions
	//url := strings.Replace(m.Announce, "announce", "scrape", 1)
	//request := url + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + GenPeerId() +
	request := m.Announce + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + GenPeerId() +
		"&uploaded=0" +
		"&downloaded=0" +
		"&left=" + strconv.Itoa(int(m.Info.Length)) +
		"&port=51413" +
		"&key=60502143" + "&numwant=80&compact=1&supportcrypto=1" +
		"&event=started"

	// TODO: conStruct
	fmt.Println("GET:", request)
	resp, err := http.Get(request)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close() // Body is a ReadCloser

	//fmt.Println(resp.Body)
	// Decode bencoded Response
	dec := bencode.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	//fmt.Println(string(resp.Body))
	if err != nil {
		return response, err
	}
	//reader := bytes.NewReader(response.Peers)
	//dec = bencode.NewDecoder(reader)
	//dec.Decode(&response.PeerDict)
	//fmt.Println(response.PeerDict)

	return response, err

}

func GenPeerId() string {
	b := [20]byte{'-', 'T', 'R', '3', '4', '4', '0', '-',
		'9', '1', 'a', '2', '4', 'W', '5', '7', '7', '4', '6', '1'}
	return UrlEncode(b)
}
