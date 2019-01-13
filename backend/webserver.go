package backend

import (
	"Peerster/entities"
	"Peerster/frontend"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// Used to make the gossiper visible from everywhere in the file
var gossiper *entities.Gossiper

// Used to make the TID channel from everywhere in the file
var idChannel *chan uint32

// Webserver - Lauch a webserver on port 8080
func Webserver(g *entities.Gossiper, chanID chan uint32) {

	// Make the gossiper and channel visible
	gossiper = g
	idChannel = &chanID

	r := mux.NewRouter()

	// POST requests
	r.HandleFunc("/rumor", postRumorHandler).Methods("POST")
	r.HandleFunc("/private", postPrivateHandler).Methods("POST")
	r.HandleFunc("/node", postNodeHandler).Methods("POST")
	r.HandleFunc("/fileIndex", postFileIndexHandler).Methods("POST")
	r.HandleFunc("/fileRequest", postFileRequestMonoSourceHandler).Methods("POST")
	r.HandleFunc("/fileRequestNetwork", postFileRequestMultiSourceHandler).Methods("POST")
	r.HandleFunc("/fileSearch", postFileSearchHandler).Methods("POST")
	r.HandleFunc("/private", postPrivateHandler).Methods("POST")

	// ArtSystem
	r.HandleFunc("/subscribe", postSubscribeHandler).Methods("POST")
	r.HandleFunc("/download", getDownloadHandler).Methods("GET")

	// Updates
	r.HandleFunc("/updates", getUpdatesHandler).Methods("GET")

	// ID
	r.HandleFunc("/id", getIDHandler).Methods("GET")

	// Root page
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:" + g.Args.ServerPort,
	}

	// Launch the server
	srv.ListenAndServe()
}

// ConfirmAndParse - Parses the received JSON and confirms reception to the frotnend
func ConfirmAndParse(w http.ResponseWriter, r *http.Request) *map[string]interface{} {

	// Confirm POST to frontend
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	// Parse received JSON
	var recJSON map[string]interface{}
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		if err := json.Unmarshal(data, &recJSON); err != nil {
			return nil
		}
	} else {
		return nil
	}

	return &recJSON
}

func getIDHandler(w http.ResponseWriter, r *http.Request) {

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(map[string]string{
		"name": gossiper.Args.Name,
		"addr": gossiper.Args.GossipAddr,
	})
	w.Write(data)

}

func getUpdatesHandler(w http.ResponseWriter, r *http.Request) {

	// Get data
	data := frontend.FBuffer.GetDataAndEmpty()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}
