package main

import (
	"log"
	"net/http"
	"spotify-auto-p/internal/session"
)

func main() {
	PORT := ":8888"
	session.SessionMap
	http.HandleFunc("callback", AuthHandler)
	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Println("Error starting server")
	}
}
