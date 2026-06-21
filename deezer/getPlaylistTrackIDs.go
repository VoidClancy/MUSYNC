package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type playlistTracksResponse struct {
	Results struct {
		Songs struct {
			Data []struct {
				SngID string `json:"SNG_ID"`
			} `json:"data"`
			Total int `json:"total"`
		} `json:"SONGS"`
	} `json:"results"`
}

func GetPlaylistTrackIDs(session *Session, playlistID int64) (map[int64]bool, error) {
	url := fmt.Sprintf(
		"https://www.deezer.com/ajax/gw-light.php?method=deezer.pagePlaylist&input=3&api_version=1.0&api_token=%s",
		session.Token,
	)
	payload, _ := json.Marshal(map[string]any{
		"playlist_id": playlistID,
		"lang":        "en",
		"nb":          2000,
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", "arl="+session.ARL)
	req.Header.Set("Content-Type", "application/json")

	resp, err := session.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result playlistTracksResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w (body: %s)", err, string(body))
	}

	existing := make(map[int64]bool, len(result.Results.Songs.Data))
	for _, s := range result.Results.Songs.Data {
		id, err := strconv.ParseInt(s.SngID, 10, 64)
		if err != nil {
			continue
		}
		existing[id] = true
	}
	return existing, nil
}
