package backend

import (
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
		return // Ignore
	}

	// Serve the file
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	http.ServeFile(w, r, dir+"\\_Downloads\\"+filename)
}
