package backend

import (
	"Peerster/network"
	"net/http"
)

func postSubscribeHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	signature, ok := (*recJSON)["signature"].(string)
	if ok {
		return // Ignore
	}

	network.OnSubscribe(gossiper, signature)
}

func getDownloadHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	filename, ok := (*recJSON)["filename"].(string)
	if ok {
		return // Ignore
	}

	// Serve the file
	http.ServeFile(w, r, filename)
}
