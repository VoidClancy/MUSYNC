package main

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"musync/logger"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	deezerArl := os.Getenv("DEEZER_ARL")

	if deezerArl == "" {
		logger.Error("DEEZER_ARL environment variable is not set")
		os.Exit(1)
	}

	PLAYLISTS := GetPlaylists()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	runSync(client, deezerArl, PLAYLISTS)
}
