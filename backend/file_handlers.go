package backend

import (
	"Peerster/files"
	"Peerster/messages"
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
	if file := gossiper.FileIndex.AddLocalFile(filename); file != nil {
		// Broadcast the transaction and publish to the blockchain
		network.OnReceiveTransaction(gossiper, &messages.TxPublish{
			File:     file,
			HopLimit: network.TransactionHopLimit + 1,
		}, nil)
	}

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

	if budget, err := strconv.ParseInt(budgetStr, 10, 32); err == nil { // Parse budget
		// Initiate file search
		network.OnInitiateFileSearch(gossiper, uint64(budget), strings.Split(keywords, ","))
	}

}
