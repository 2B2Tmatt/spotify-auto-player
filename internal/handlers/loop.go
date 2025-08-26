package handlers

import (
	"log"
	"net/http"
	sessions "spotify-auto-p/internal/session"
	"time"
)

func Loop(w http.ResponseWriter, r *http.Request, store *sessions.Store) {
	pauses := 0
	client := http.Client{Timeout: time.Second * 8}
	stateReq, err := http.NewRequest(http.MethodGet, "https://api.spotify.com/v1/me/player", nil)
	stateReq.Header.Set("Authorization")
	if err != nil {
		log.Println(err)
		return
	}

	for pauses < 5 {

	}
}
