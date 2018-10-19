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

	// /message subdomain
	r.HandleFunc("/message", postMessageHandler).Methods("POST")
	r.HandleFunc("/message", getMessageHandler).Methods("GET")

	// /node subdomain
	r.HandleFunc("/node", postNodeHandler).Methods("POST")
	r.HandleFunc("/node", getNodeHandler).Methods("GET")

	// /id subdomain
	r.HandleFunc("/id", getIDHandler).Methods("GET")

	// Initialization
	r.HandleFunc("/in_message", getInitMessageHandler).Methods("GET")
	r.HandleFunc("/in_node", getInitNodeHandler).Methods("GET")

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
	})
	w.Write(data)

}
