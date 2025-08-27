package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"spotify-auto-p/config"
	"spotify-auto-p/internal/handlers"
	sessions "spotify-auto-p/internal/session"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	PORT := ":8888"
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with system env")
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
	http.HandleFunc("/cleanse", func(w http.ResponseWriter, r *http.Request) { handlers.Cleanse(w, r, store) })
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser("http://127.0.0.1:8888/auth")
	}()
	err = http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Println("Error starting server")
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}
