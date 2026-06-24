package deezer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DeleteSongsFromPlaylist(session *Session, playlistID int64, trackIDs []int64) error {
	url := fmt.Sprintf(
		"https://www.deezer.com/ajax/gw-light.php?method=playlist.deleteSongs&input=3&api_version=1.0&api_token=%s",
		session.Token,
	)

	songs := make([][]int64, len(trackIDs))
	for i, id := range trackIDs {
		songs[i] = []int64{id, int64(i)}
	}

	payload, err := json.Marshal(map[string]any{
		"playlist_id": playlistID,
		"songs":       songs,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "arl="+session.ARL)
	req.Header.Set("Content-Type", "application/json")

	resp, err := session.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Error map[string]any `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err == nil && len(result.Error) > 0 {
		return fmt.Errorf("deezer api error: %v (raw response: %s)", result.Error, string(body))
	}

	return nil
}
