package backend

import (
	"Peerster/network"
	"Peerster/types"
	"net/http"
)

func postRumorHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
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

	recJSON := ConfirmAndParse(w, r)
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
