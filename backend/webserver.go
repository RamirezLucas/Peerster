package backend

import (
	"Peerster/network"
	"Peerster/types"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// Used to make the gossiper visible from everywhere in the file
var gossiper *types.Gossiper

// Used to make the TID channel from everywhere in the file
var idChannel *chan uint32

// Webserver - Lauch a webserver on port 8080
func Webserver(g *types.Gossiper, chanID chan uint32) {

	// Make the gossiper and channel visible
	gossiper = g
	idChannel = &chanID

	r := mux.NewRouter()

	// /message subdomain
	r.HandleFunc("/message", postMessageHandler).Methods("POST")
	r.HandleFunc("/message", getMessageHandler).Methods("GET")

	// /node subdomain
	r.HandleFunc("/node", postNodeHandler).Methods("POST")
	r.HandleFunc("/node", getNodeHandler).Methods("GET")

	// /id subdomain
	r.HandleFunc("/id", getIDHandler).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
	}

	// Launch the server
	srv.ListenAndServe()
}

func postMessageHandler(w http.ResponseWriter, r *http.Request) {

	// Confirm POST to frontend
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	// Parse received JSON
	var newMsg map[string]interface{}
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, &newMsg); err != nil {
			return // Ignore
		}
	} else {
		return // Ignore
	}

	// Typecheck
	msg, ok := newMsg["msg"].(string)
	if !ok {
		return // Ignore
	}

	// Accep the new message
	rumor := &types.RumorMessage{Text: msg}
	network.OnReceiveRumor(gossiper, rumor, nil, true, <-(*idChannel))

}

func getMessageHandler(w http.ResponseWriter, r *http.Request) {

	// Get data
	data := types.BufferMessages.GetDataAndEmpty()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

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
	w.Write(data)

}

func getIDHandler(w http.ResponseWriter, r *http.Request) {

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(map[string]string{
		"name": gossiper.Name,
	})
	w.Write(data)

}
