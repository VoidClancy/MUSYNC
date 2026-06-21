package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type sendMessagePayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type Audit struct {
	Name           string
	ISRCsFound     int
	TracksResolved int
	TracksMissed   int
	AlreadyPresent int
	TracksAdded    int
	Err            error
}

func SendSyncSummary(botToken, chatID string, duration time.Duration, audits []Audit) error {
	if botToken == "" || chatID == "" {
		return nil
	}

	var totalISRCs, totalResolved, totalAdded, totalErrors int
	var detailsBuf bytes.Buffer

	for _, a := range audits {
		totalISRCs += a.ISRCsFound
		totalResolved += a.TracksResolved
		totalAdded += a.TracksAdded

		statusIcon := "[ OK ]"
		statusText := "OK"
		if a.Err != nil {
			statusIcon = "[FAIL]"
			statusText = "FAILED: " + a.Err.Error()
			totalErrors++
		}

		fmt.Fprintf(&detailsBuf, "%s *%s*\n", statusIcon, a.Name)
		fmt.Fprintf(&detailsBuf, "• Tracks: %d | Resolved: %d\n", a.ISRCsFound, a.TracksResolved)
		if a.TracksAdded > 0 {
			fmt.Fprintf(&detailsBuf, "• Added: %d new tracks\n", a.TracksAdded)
		}
		if a.Err != nil {
			fmt.Fprintf(&detailsBuf, "• Status: %s\n", statusText)
		}
		detailsBuf.WriteString("\n")
	}

	summaryHeader := "=== Deezer Sync Summary ==="
	if totalErrors > 0 {
		summaryHeader = "=== Deezer Sync Summary (Errors Encountered) ==="
	}

	summaryText := fmt.Sprintf(
		"*%s*\n\n"+
			"*Duration:* %s\n"+
			"*Playlists:* %d total (%d failed)\n"+
			"*Tracks:* %d found | %d resolved | %d added\n\n"+
			"%s",
		summaryHeader,
		duration.Round(time.Second),
		len(audits),
		totalErrors,
		totalISRCs,
		totalResolved,
		totalAdded,
		detailsBuf.String(),
	)

	payload := sendMessagePayload{
		ChatID:    chatID,
		Text:      summaryText,
		ParseMode: "Markdown",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %s", resp.Status)
	}

	return nil
}
