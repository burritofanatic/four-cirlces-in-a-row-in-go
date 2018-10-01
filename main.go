package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT must be set! Defaulting to 8080")
		port = "8080"
	}

	flag.Parse()
	log.SetFlags(0)
	hub := NewHub()
	router := mux.NewRouter()

	// Initial Join
	router.HandleFunc("/ws/{deviceId}", hub.HandleWS).Methods("GET")

	// Move
	router.HandleFunc("/move", hub.HandleMove).Methods("POST")

	// Resign
	router.HandleFunc("/resign", hub.HandleResignRequest).Methods("POST")

	// Leave Room
	router.HandleFunc("/leave", hub.HandleLeaveRequest).Methods("POST")

	// Chat message
	router.HandleFunc("/message", hub.HandleMessage).Methods("POST")

	// Rematch Request
	router.HandleFunc("/rematch", hub.HandleRematchRequest).Methods("POST")

	// Find Games
	router.HandleFunc("/find", hub.HandleFindAllGamesPlayedRequest).Methods("POST")

	http.Handle("/", router)

	// Print, Listen, and Serve
	log.Printf("http_err: %v", http.ListenAndServe(":"+port, nil))
}
