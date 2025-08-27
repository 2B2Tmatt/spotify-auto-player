package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type PlaybackState struct {
	IsPlaying    bool    `json:"is_playing"`
	Item         Track   `json:"item"`
	TrackContext Context `json:"context"`
}

type Context struct {
	Type        string `json:"type"`
	PlaylistURL string `json:"href"`
}

type Track struct {
	Explicit bool   `json:"explicit"`
	ID       string `json:"id"`
	URI      string `json:"uri"`
}

type User struct {
	SessionToken string
	Client       *http.Client
}

type Playlist struct {
	PlaylistTracks Tracks `json:"tracks"`
}

type Tracks struct {
	PlaylistItems []Items `json:"items"`
}

type Items struct {
	PlaylistTracks Track `json:"track"`
}

type TracksEncode struct {
	TracksToEncode []TrackToEncode `json:"tracks"`
}

type TrackToEncode struct {
	URI string `json:"uri"`
}

func (user *User) RemoveSongsFromPlaylist(state PlaybackState, songs TracksEncode) (bool, error) {
	jsonData, err := json.Marshal(songs)
	if err != nil {
		return false, err
	}
	bodyReader := bytes.NewBuffer(jsonData)
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/tracks", state.TrackContext.PlaylistURL), bodyReader)
	if err != nil {
		log.Println(err)
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	return true, nil
}

func (user *User) GetPlaylist(state PlaybackState) (Playlist, error) {
	req, err := http.NewRequest(http.MethodGet, state.TrackContext.PlaylistURL, nil)
	if err != nil {
		log.Println(err)
		return Playlist{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return Playlist{}, err
	}
	defer resp.Body.Close()
	var playlist Playlist
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return Playlist{}, err
	}
	log.Println(resp.StatusCode)
	return playlist, nil
}

func (user *User) GetPlaybackState() (PlaybackState, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.spotify.com/v1/me/player", nil)
	if err != nil {
		log.Println(err)
		return PlaybackState{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return PlaybackState{}, err
	}
	defer resp.Body.Close()
	var state PlaybackState
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return PlaybackState{}, err
	}
	log.Println(resp.StatusCode)
	return state, nil
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

func (user *User) Play() error {
	req, err := http.NewRequest(http.MethodPut, "https://api.spotify.com/v1/me/player/play", nil)
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
	log.Println("Play:", resp.StatusCode)
	return nil
}

func (user *User) CheckExplicit(track Track) (bool, error) {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", track.ID), nil)
	if err != nil {
		log.Println(err)
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", user.SessionToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := user.Client.Do(req)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&track)
	log.Println("Check explicit:", resp.StatusCode)
	return track.Explicit, nil
}
