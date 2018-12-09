package backend

import (
	"Peerster/fail"
	"Peerster/files"
	"Peerster/network"
	"encoding/hex"
	"net/http"
	"strconv"
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

func postFileRequestMonoSourceHandler(w http.ResponseWriter, r *http.Request) {

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

func postFileRequestMultiSourceHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	filename, ok1 := (*recJSON)["filename"].(string)
	metahash, ok2 := (*recJSON)["metahash"].(string)
	if !ok1 || !ok2 {
		return // Ignore
	}

	fail.LeveledPrint(1, "postFileRequestMultiSourceHandler", "Request for %s with metahash %s", filename, metahash)

	if decoded, err := hex.DecodeString(metahash); err == nil && len(decoded) == files.HashSizeBytes {
		// Starts file reconstruction
		network.OnRemoteMetafileRequestMultisource(gossiper, decoded, filename)
	}
}

func postFileSearchHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	budgetStr, ok1 := (*recJSON)["budget"].(string)
	keywords, ok2 := (*recJSON)["keywords"].(string)
	if !ok1 || !ok2 {
		return // Ignore
	}

	fail.LeveledPrint(1, "postFileSearchHandler", "Request for %s with budget %s", keywords, budgetStr)

	if budget, err := strconv.ParseInt(budgetStr, 10, 32); err == nil { // Parse budget
		// Initiate file search
		network.OnInitiateFileSearch(gossiper, uint64(budget), strings.Split(keywords, ","))
	}

}
