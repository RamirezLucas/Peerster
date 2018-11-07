package backend

import (
	"Peerster/network"
	"Peerster/types"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func confirmAndParse(w http.ResponseWriter, r *http.Request) *map[string]interface{} {

	// Confirm POST to frontend
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	// Parse received JSON
	var recJSON map[string]interface{}
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, &recJSON); err != nil {
			fmt.Printf("a")
			return nil
		}
	} else {
		fmt.Printf("b")
		return nil
	}

	return &recJSON
}

func postRumorHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := confirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	msg, ok := (*recJSON)["message"].(string)
	if !ok {
		return // Ignore
	}

	// Accept the new message
	rumor := &types.RumorMessage{Text: msg}
	network.OnReceiveClientRumor(gossiper, rumor, <-(*idChannel))

}

func postPrivateHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := confirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	dst, ok1 := (*recJSON)["destination"].(string)
	msg, ok2 := (*recJSON)["message"].(string)
	if !ok1 || !ok2 {
		return // Ignore
	}

	// Accept the new message
	privateMessage := &types.PrivateMessage{Destination: dst, Text: msg}
	network.OnReceiveClientPrivate(gossiper, privateMessage)
}

func getUpdatesHandler(w http.ResponseWriter, r *http.Request) {

	// Get data
	data := types.FBuffer.GetDataAndEmpty()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}

// func getInitMessageHandler(w http.ResponseWriter, r *http.Request) {

// 	data := gossiper.NameIndex.GetEverything()

// 	// Send JSON data
// 	w.WriteHeader(http.StatusOK)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(*data)

// }
