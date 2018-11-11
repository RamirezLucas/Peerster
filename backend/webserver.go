package backend

import (
	"Peerster/types"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// Used to make the gossiper visible from everywhere in the file
var gossiper *types.Gossiper

// Used to make the TID channel from everywhere in the file
var idChannel *chan uint32

// Webserver - Lauch a webserver on port 8080
func Webserver(g *types.Gossiper, chanID chan uint32) {

	// Make the gossiper and channel visible
	gossiper = g
	idChannel = &chanID

	r := mux.NewRouter()

	// POST requests
	r.HandleFunc("/rumor", postRumorHandler).Methods("POST")
	r.HandleFunc("/private", postPrivateHandler).Methods("POST")
	r.HandleFunc("/node", postNodeHandler).Methods("POST")
	r.HandleFunc("/file_index", postPrivateHandler).Methods("POST")
	r.HandleFunc("/file_request", postPrivateHandler).Methods("POST")
	r.HandleFunc("/private", postPrivateHandler).Methods("POST")

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
	data := types.FBuffer.GetDataAndEmpty()

	// Send JSON data
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(*data)

}
