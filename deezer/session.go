package deezer

import "net/http"

// Session holds the shared connection and authentication context for Deezer API operations.
type Session struct {
	HTTPClient *http.Client
	ARL        string
	Token      string
	UserID     int64
}
