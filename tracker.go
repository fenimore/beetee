package main

import "github.com/anacrolix/torrent/bencode"
import "strconv"
import "net/http"

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
	FailureReason string `bencode:"failure reason"`
	Interval      int32  `bencode:"interval"`
	//	IntervalMin int64         `bencode:"min interval"`
	TrackerId  string        `bencode:"tracker id"`
	Complete   int32         `bencode:"complete"`
	Incomplete int32         `bencode:"incomplete"`
	Peers      bencode.Bytes `bencode:"peers"`
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
	request := m.Announce + "?info_hash=" + m.InfoHashEnc + "&peer_id=" + UrlEncode(PeerId) +
		"&uploaded=0" +
		"&downloaded=0" +
		"&left=" + strconv.Itoa(int(m.Info.Length)) +
		"&port=6882" +
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

	response.parsePeers()

	return response, err
}

func GenPeerId() [20]byte {
	// TODO: Make random
	b := [20]byte{'-', 'F', 'L', '1', '0', '0', '1', '-',
		'9', '1', 'a', '2', '4', 'W', '5', '7', '7', '4', '6', '1'}
	return b
}
