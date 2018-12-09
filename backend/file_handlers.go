package backend

import (
	"Peerster/files"
	"Peerster/network"
	"encoding/hex"
	"net/http"
	"strings"
)

func postFileIndexHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	filename, ok := (*recJSON)["filename"].(string)
	if !ok {
		return // Ignore
	}

	// Index the new file
	gossiper.FileIndex.AddLocalFile(filename)

}

func postFileRequestHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	filename, ok1 := (*recJSON)["filename"].(string)
	metahash, ok2 := (*recJSON)["metahash"].(string)
	destination, ok3 := (*recJSON)["destination"].(string)
	if !ok1 || !ok2 || !ok3 {
		return // Ignore
	}

	if decoded, err := hex.DecodeString(metahash); err == nil && len(decoded) == files.HashSizeBytes {
		// Starts file reconstruction
		network.OnRemoteMetafileRequestMonosource(gossiper, decoded, filename, destination)
	}
}

func postFileSearchHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	budget, ok1 := (*recJSON)["budget"].(uint64)
	keywords, ok2 := (*recJSON)["keywords"].(string)
	if !ok1 || !ok2 {
		return // Ignore
	}

	// Handle particular case of budget == 0
	if budget == 0 {
		budget = ^uint64(0)
	}

	// Initiate file search
	network.OnInitiateFileSearch(gossiper, budget, strings.Split(keywords, ","))
}
