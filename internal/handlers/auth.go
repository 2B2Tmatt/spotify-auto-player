package handlers

import (
	"log"
	"net/http"
	"spotify-auto-p/config"
	"spotify-auto-p/internal/auth"
	sessions "spotify-auto-p/internal/session"
	"time"
)

func AuthStart(w http.ResponseWriter, r *http.Request, store *sessions.Store, cfg *config.Config) {
	details := cfg
	state, err := auth.GenerateRandomString(32)
	if err != nil {
		log.Println("error generating state", err)
		return
	}
	verifier, code_challenge, err := auth.NewPKCE()
	if err != nil {
		log.Println("error generating new PKCE", err)
		return
	}
	sid := store.EnsureSessionID(w, r)
	err = store.SetPendingAuth(sid, sessions.PendingAuth{
		State:        state,
		CodeVerifier: verifier,
		Scopes:       details.Scopes,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		log.Println(err)
		return
	}
	auth.RequestUserAuth(w, r, details.ClientID, details.RedirectURI, state, code_challenge, details.Scopes)
}
