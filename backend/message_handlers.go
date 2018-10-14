package backend

import (
	"Peerster/network"
	"Peerster/types"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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
	w.Write(*data)

}

func getInitMessageHandler(w http.ResponseWriter, r *http.Request) {

	data := gossiper.NameIndex.GetEverything()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}
