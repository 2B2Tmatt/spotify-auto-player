package auth

import (
	"net/http"
	"net/url"
	"strings"
)

func RequestUserAuth(
	w http.ResponseWriter,
	r *http.Request,
	clientID, redirectURI, state, codeChallenge string,
	scopes []string,
) {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("code_challenge_method", "S256")
	q.Set("code_challenge", codeChallenge)
	q.Set("state", state)

	u := &url.URL{
		Scheme:   "https",
		Host:     "accounts.spotify.com",
		Path:     "/authorize",
		RawQuery: q.Encode(),
	}
	http.Redirect(w, r, u.String(), http.StatusFound)
}
