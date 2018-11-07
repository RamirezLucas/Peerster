package backend

import (
	"Peerster/types"
	"encoding/json"
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

	// POST requests for new message
	r.HandleFunc("/rumor", postRumorHandler).Methods("POST")
	r.HandleFunc("/private", postPrivateHandler).Methods("POST")

	// /updates subdomain
	r.HandleFunc("/updates", getUpdatesHandler).Methods("GET")

	// /node subdomain
	r.HandleFunc("/node", postNodeHandler).Methods("POST")

	// /id subdomain
	r.HandleFunc("/id", getIDHandler).Methods("GET")

	// Initialization
	// r.HandleFunc("/in_message", getInitMessageHandler).Methods("GET")
	// r.HandleFunc("/in_node", getInitNodeHandler).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/")))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:" + g.Args.ServerPort,
	}

	// Launch the server
	srv.ListenAndServe()
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
