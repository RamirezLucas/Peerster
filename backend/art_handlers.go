package backend

import (
	"Peerster/fail"
	"Peerster/network"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func postSubscribeHandler(w http.ResponseWriter, r *http.Request) {

	recJSON := ConfirmAndParse(w, r)
	if recJSON == nil {
		return // Ignore
	}

	// Typecheck
	signature, ok := (*recJSON)["signature"].(string)
	if !ok {
		return // Ignore
	}

	network.OnSubscribe(gossiper, signature)
}

func postDownloadHandler(w http.ResponseWriter, r *http.Request) {

	fail.LeveledPrint(1, "getDownloadHandler", "Received download request")

	// recJSON := ConfirmAndParse(w, r)
	// if recJSON == nil {
	// 	fail.LeveledPrint(1, "getDownloadHandler", "Error parsing")
	// 	return // Ignore
	// }

	var recJSON map[string]interface{}
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, &recJSON); err != nil {
			return
		}
	} else {
		return
	}

	// Typecheck
	filename, ok := recJSON["filename"].(string)
	if !ok {
		fail.LeveledPrint(1, "getDownloadHandler", "Error")
		return // Ignore
	}

	// Serve the file
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fail.LeveledPrint(1, "getDownloadHandler", "Uploading %s", dir+"\\_Downloads\\"+filename)
	http.ServeFile(w, r, dir+"\\_Downloads\\"+filename)
}
