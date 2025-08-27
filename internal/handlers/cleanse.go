package handlers

import (
	"log"
	"net/http"
	sessions "spotify-auto-p/internal/session"
	"spotify-auto-p/internal/spotify"
	"time"
)

func Cleanse(w http.ResponseWriter, r *http.Request, store *sessions.Store) {
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
	state, err := user.GetPlaybackState()
	if err != nil {
		log.Println("Error getting playback state", err)
	}
	log.Println("This is the state: ", state)
	if state.TrackContext.Type != "playlist" {
		log.Println("Not currently on a playlist")
		http.Error(w, "Not currently on a playlist", http.StatusAccepted)
		return
	}
	playlist, err := user.GetPlaylist(state)
	if err != nil {
		log.Println("Error getting playlist state", err)
	}
	log.Println("This is the playlist", playlist.PlaylistTracks.PlaylistItems[0].PlaylistTracks)
	var deletions spotify.TracksEncode
	for _, song := range playlist.PlaylistTracks.PlaylistItems {
		if explicit, _ := user.CheckExplicit(song.PlaylistTracks); explicit {
			deletion := spotify.TrackToEncode{URI: song.PlaylistTracks.URI}
			deletions.TracksToEncode = append(deletions.TracksToEncode, deletion)
		}
	}
	_, err = user.RemoveSongsFromPlaylist(state, deletions)
	if err != nil {
		log.Println("Error removing songs", err)
		return
	}
	http.Error(w, "Playlist Clean", http.StatusAccepted)
}
