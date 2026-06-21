package main

import (
	"encoding/json"
	"os"
	"musync/logger"
)

type Playlist struct {
	Name     string `json:"Name"`
	Url      string `json:"-"`
	Id       string `json:"Id"`
	DeezerID int64  `json:"DeezerID"`
}

func (p *Playlist) updateDeezerID(id int64) {
	p.DeezerID = id
}

func GetPlaylists() []Playlist {
	var PLAYLISTS []Playlist

	file, err := os.ReadFile("playlists.json")
	if err != nil {
		logger.Error("can't find playlists file")
		os.Exit(1)
	}

	err = json.Unmarshal(file, &PLAYLISTS)
	if err != nil {
		logger.Error("unmarshal playlists failed", "err", err)
		os.Exit(1)
	}

	for i := range PLAYLISTS {
		PLAYLISTS[i].Url = "https://isrchunt.com/?spotifyPlaylist=https%3A%2F%2Fopen.spotify.com%2Fplaylist%2F" + PLAYLISTS[i].Id
	}

	return PLAYLISTS
}

func SavePlaylists(playlists []Playlist) error {
	data, err := json.MarshalIndent(playlists, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("playlists.json", data, 0644)
}
