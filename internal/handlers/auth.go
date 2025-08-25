package handlers

import (
	"log"
	"net/http"
	"net/url"
	"spotify-auto-p/config"
	"spotify-auto-p/internal/auth"
	sessions "spotify-auto-p/internal/session"
	"time"
)

func AuthStart(w http.ResponseWriter, r *http.Request, store *sessions.Store) {
	details, err := config.FromEnv()
	if err != nil {
		log.Println("Error getting details from config/env", err)
		return
	}
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
	store.AddPendingAuth(sid, &sessions.PendingAuth{
		State:        state,
		CodeVerifier: verifier,
		Scopes:       details.Scopes,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	})
	auth.RequestUserAuth(w, r, details.ClientID, details.RedirectURI, state, code_challenge, details.Scopes)
}

func AuthCallBack(w http.ResponseWriter, r *http.Request, store *sessions.Store) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		log.Println("No cookie in request", err)
		return
	}
	sid := cookie.Value
	session, exists := store.Get(sid)
	if !exists {
		log.Println("No session exists", err)
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	urlerr := r.URL.Query().Get("error")
	if urlerr != "" {
		if state == session.PendingAuth.State {
			store.Delete(sid)
			log.Println("Error Authenticating")
			return
		}
		log.Println("Use the offical start point")
		return
	}
	if state != session.PendingAuth.State {
		store.Delete(sid)
		log.Println("Error Authenticating")
		return
	}
	if session.PendingAuth.ExpiresAt.Before(time.Now()) {
		store.Delete(sid)
		log.Println("Session ran out of time")
		return
	}
	if code == "" {
		store.Delete(sid)
		log.Println("Invalid callback")
		return
	}
	details, err := config.FromEnv()
	if err != nil {
		log.Println("Error getting details from config/env", err)
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	q := url.Values{}
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("redirect_uri", details.RedirectURI)
	q.Set("client_id", details.ClientID)
	q.Set("code_verifier", session.PendingAuth.CodeVerifier)
	u := &url.URL{
		Scheme:   "https",
		Host:     "accounts.spotify.com",
		Path:     "/api/token",
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error executing request")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Error getting response token")
		return
	}
	var token http.Tokens

}
