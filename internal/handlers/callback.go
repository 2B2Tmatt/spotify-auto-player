package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"spotify-auto-p/config"
	sessions "spotify-auto-p/internal/session"
	"strings"
	"time"
)

func AuthCallBack(w http.ResponseWriter, r *http.Request, store *sessions.Store, cfg *config.Config) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		log.Println("No cookie in request", err)
		return
	}
	sid := cookie.Value
	_, exists := store.GetSessionSnapshot(sid)
	if !exists {
		log.Println("AuthCallBack - No session exists", err)
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	urlerr := r.URL.Query().Get("error")
	pendingAuth, exists, err := store.GetPendingAuthCopy(sid)
	if !exists {
		log.Println("Error in previous Auth stage", err)
		return
	}

	if urlerr != "" {
		if state == pendingAuth.State {
			store.RemovePendingAuth(sid)
			log.Println("Error Authenticating")
			return
		}
		log.Println("Use the offical start point")
		return
	}
	if state != pendingAuth.State {
		store.RemovePendingAuth(sid)
		log.Println("Error Authenticating")
		return
	}
	if pendingAuth.ExpiresAt.Before(time.Now()) {
		store.RemovePendingAuth(sid)
		log.Println("Session ran out of time")
		return
	}
	if code == "" {
		store.RemovePendingAuth(sid)
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
		"code_verifier": []string{pendingAuth.CodeVerifier},
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
	err = store.SetSessionToken(sid, token)
	if err != nil {
		log.Println(err)
		return
	}
	tokenCopy, exists, err := store.GetTokensCopy(sid)
	if !exists {
		log.Println(err)
		return
	}

	log.Println("I worked!")
	log.Println("Session token:", tokenCopy.AccessToken, "\n\n")
	store.RemovePendingAuth(sid)

	http.Redirect(w, r, "/loop", http.StatusSeeOther)
}
