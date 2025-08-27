package spotify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type PlaybackState struct {
	IsPlaying bool `json:"is_playing"`
}

type User struct {
	SessionToken string
	Client       *http.Client
}

func (user *User) GetPlaybackState() (*PlaybackState, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.spotify.com/v1/me/player", nil)
	if err != nil {
		log.Println(err)
		return &PlaybackState{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return &PlaybackState{}, err
	}
	defer resp.Body.Close()
	var state PlaybackState
	json.NewDecoder(resp.Body).Decode(&state)
	log.Println(resp.StatusCode)
	return &state, nil
}

func (user *User) Pause() error {
	req, err := http.NewRequest(http.MethodPut, "https://api.spotify.com/v1/me/player/pause", nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	return nil
}
