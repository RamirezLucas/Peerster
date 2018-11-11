package backend

import (
	"Peerster/network"
	"Peerster/types"
	"encoding/hex"
	"net/http"
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
	gossiper.FileIndex.IndexNewFile(filename)

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

	if decoded, err := hex.DecodeString(metahash); err == nil && len(decoded) == types.HashSizeBytes {
		// Starts file reconstruction
		network.OnRemoteMetaFileRequest(gossiper, decoded, filename, destination)
	}
}