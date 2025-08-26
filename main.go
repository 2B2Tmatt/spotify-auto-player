package main

import (
	"log"
	"net/http"
	"os"
	"spotify-auto-p/config"
	"spotify-auto-p/internal/handlers"
	sessions "spotify-auto-p/internal/session"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	PORT := ":8888"
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with system env")
		return
	}
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	Cfg := config.Config{
		ClientID:    clientID,
		RedirectURI: redirectURI,
	}
	ScopesRaw := strings.TrimSpace(os.Getenv("SPOTIFY_SCOPES"))
	if ScopesRaw != "" {
		Cfg.Scopes = strings.Fields(ScopesRaw)
	}
	if Cfg.ClientID == "" || Cfg.RedirectURI == "" {
		log.Printf("ClientID:%s RedirectURI:%s", Cfg.ClientID, Cfg.RedirectURI)
		return
	}

	store := sessions.NewStore()
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { handlers.AuthStart(w, r, store, &Cfg) })
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) { handlers.AuthCallBack(w, r, store, &Cfg) })
	http.HandleFunc("/loop", handlers.Loop)
	err = http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Println("Error starting server")
	}
}
