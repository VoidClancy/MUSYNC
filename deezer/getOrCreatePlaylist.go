package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"musync/logger"
	"strconv"
)

type playlistListResponse struct {
	Results struct {
		Tab struct {
			Playlists struct {
				Data []struct {
					ID    string `json:"PLAYLIST_ID"`
					Title string `json:"TITLE"`
				} `json:"data"`
			} `json:"playlists"`
		} `json:"TAB"`
	} `json:"results"`
}

// GetOrCreatePlaylist returns the ID of a playlist matching name,
// creating it if no match exists.
func GetOrCreatePlaylist(session *Session, name string) (int64, error) {
	id, err := findPlaylistByName(session, name)
	if err != nil {
		return 0, fmt.Errorf("listing playlists: %w", err)
	}
	if id != 0 {
		return id, nil
	}
	return createPlaylist(session, name)
}

func findPlaylistByName(session *Session, name string) (int64, error) {
	url := fmt.Sprintf(
		"https://www.deezer.com/ajax/gw-light.php?method=deezer.pageProfile&input=3&api_version=1.0&api_token=%s",
		session.Token,
	)
	payload, _ := json.Marshal(map[string]any{
		"tab":     "playlists",
		"user_id": session.UserID,
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cookie", "arl="+session.ARL)
	req.Header.Set("Content-Type", "application/json")

	resp, err := session.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var pl playlistListResponse
	if err := json.Unmarshal(body, &pl); err != nil {
		return 0, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}

	for _, p := range pl.Results.Tab.Playlists.Data {
		if p.Title == name {
			id, err := strconv.ParseInt(p.ID, 10, 64)
			if err != nil {
				continue
			}
			return id, nil
		}
	}
	return 0, nil
}

func createPlaylist(session *Session, name string) (int64, error) {
	url := fmt.Sprintf(
		"https://www.deezer.com/ajax/gw-light.php?method=playlist.create&input=3&api_version=1.0&api_token=%s",
		session.Token,
	)

	payload, _ := json.Marshal(map[string]any{
		"title":       name,
		"songs":       [][]int64{},
		"status":      0, // private; 1 = public, unverified
		"description": "",
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cookie", "arl="+session.ARL)
	req.Header.Set("Content-Type", "application/json")

	resp, err := session.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Results int64           `json:"results"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}

	// error is "[]" when empty, otherwise a populated object
	if len(result.Error) > 2 { // more than just "[]"
		return 0, fmt.Errorf("deezer api error: %s", string(result.Error))
	}

	if result.Results == 0 {
		return 0, fmt.Errorf("create returned no playlist id (body: %s)", string(body))
	}

	logger.Infof("[CREATED PLAYLIST] %s id: %d", name, result.Results)
	return result.Results, nil
}
