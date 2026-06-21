package deezer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"musync/logger"
	"strings"
)

type gatewayResponse struct {
	Results struct {
		CheckForm string `json:"checkForm"`
		User      struct {
			ID int64 `json:"USER_ID"`
		} `json:"USER"`
	} `json:"results"`
}

func GetDeezerToken(client *http.Client, arl string) (token string, userID int64, err error) {
	req, err := http.NewRequest("GET",
		"https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&api_version=1.0&api_token=",
		nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Cookie", "arl="+arl)

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var gw gatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gw); err != nil {
		return "", 0, err
	}
	if gw.Results.CheckForm == "" {
		return "", 0, fmt.Errorf("no checkForm returned, arl may be invalid or expired")
	}
	if gw.Results.User.ID == 0 {
		return "", 0, fmt.Errorf("no user_id returned in getUserData response")
	}

	logger.Infof("[OBTAINED DEEZER TOKEN]: %sxxxxx", strings.TrimSpace(gw.Results.CheckForm)[:4])
	return gw.Results.CheckForm, gw.Results.User.ID, nil
}
