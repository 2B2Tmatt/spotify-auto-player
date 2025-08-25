package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"spotify-auto-p/config"
	"spotify-auto-p/internal/auth"
	sessions "spotify-auto-p/internal/session"
	"strings"
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
	store.AddPendingAuth(w, r, sid, &sessions.PendingAuth{
		State:        state,
		CodeVerifier: verifier,
		Scopes:       details.Scopes,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	})
	auth.RequestUserAuth(w, r, details.ClientID, details.RedirectURI, state, code_challenge, details.Scopes)
}

func AuthCallBack(w http.ResponseWriter, r *http.Request, store *sessions.Store, cfg *config.Config) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		log.Println("No cookie in request", err)
		return
	}
	sid := cookie.Value
	session, exists := store.Get(sid)
	if !exists {
		log.Println("AuthCallBack - No session exists", err)
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	urlerr := r.URL.Query().Get("error")
	if session.PendingAuth == nil {
		log.Println("Error in previous Auth stage")
		return
	}
	if urlerr != "" {
		if state == session.PendingAuth.State {
			session.PendingAuth = &sessions.PendingAuth{}
			log.Println("Error Authenticating")
			return
		}
		log.Println("Use the offical start point")
		return
	}
	if state != session.PendingAuth.State {
		session.PendingAuth = &sessions.PendingAuth{}
		log.Println("Error Authenticating")
		return
	}
	if session.PendingAuth.ExpiresAt.Before(time.Now()) {
		session.PendingAuth = &sessions.PendingAuth{}
		log.Println("Session ran out of time")
		return
	}
	if code == "" {
		session.PendingAuth = &sessions.PendingAuth{}
		log.Println("Invalid callback")
		return
	}
	details := cfg
	client := &http.Client{Timeout: 10 * time.Second}
	bodyData := url.Values{
		"grant_type":    []string{"authorization_code"},
		"code":          []string{code},
		"redirect_uri":  []string{details.RedirectURI},
		"client_id":     []string{details.ClientID},
		"code_verifier": []string{session.PendingAuth.CodeVerifier},
	}
	encodedBody := bodyData.Encode()
	reader := strings.NewReader(encodedBody)
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", reader)
	if err != nil {
		log.Println("Error creating new request", err)
		return
	}
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
	var token sessions.Tokens
	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Println("Error decoding token", err)
		return
	}
	duration := time.Duration(token.ExpiresIn) * time.Second
	token.ExpiresAt = time.Now().Add(duration)
	if token.TokenType != "Bearer" {
		log.Println("Token field not valid")
		return
	}
	session.Tokens = &token
	session.PendingAuth = nil
	log.Println("I worked!")
}
