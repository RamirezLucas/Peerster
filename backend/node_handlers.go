package backend

import (
	"Peerster/types"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

func postNodeHandler(w http.ResponseWriter, r *http.Request) {

	// Confirm POST to frontend
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	// Parse received JSON
	var newPeer map[string]interface{}
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, &newPeer); err != nil {
			return // Ignore
		}
	} else {
		return // Ignore
	}

	// Typecheck
	peer, ok := newPeer["peer"].(string)
	if !ok {
		return // Ignore
	}

	// Accept the new peer
	udpAddr, err := net.ResolveUDPAddr("udp4", peer)
	if err != nil {
		return // Ignore
	}
	gossiper.PeerIndex.AddPeerIfAbsent(udpAddr)

}

func getNodeHandler(w http.ResponseWriter, r *http.Request) {

	// Get data
	data := types.BufferPeers.GetDataAndEmpty()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}

func getInitNodeHandler(w http.ResponseWriter, r *http.Request) {

	// Get data
	data := gossiper.PeerIndex.GetEverything()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}
