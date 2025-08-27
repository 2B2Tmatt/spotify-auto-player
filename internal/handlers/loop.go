package handlers

import (
	"log"
	"net/http"
	sessions "spotify-auto-p/internal/session"
	"spotify-auto-p/internal/spotify"
	"time"
)

func Loop(w http.ResponseWriter, r *http.Request, store *sessions.Store) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		log.Println("No cookie in request", err)
		return
	}
	sid := cookie.Value
	tokenCopy, _, err := store.GetTokensCopy(sid)
	if err != nil {
		log.Println("Error getting token", err)
	}
	tokenVal := tokenCopy.AccessToken
	log.Println("Token val:", tokenVal)
	client := http.Client{Timeout: time.Second * 8}
	user := spotify.User{SessionToken: tokenVal, Client: &client}

	log.Println("This worked")
	state, err := user.GetPlaybackState()
	if err != nil {
		log.Println("Error getting playback state", err)
	}
	log.Println("This is the state: ", *state)
	err = user.Pause()
	if err != nil {
		log.Println("Error unpausing", err)
	}
	http.Error(w, "This worked lol", http.StatusAccepted)
}
