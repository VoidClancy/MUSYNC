package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	TracksDeleted  int
	Err            error
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"`", "\\`",
	)
	return replacer.Replace(s)
}

func SendSyncSummary(botToken, chatID string, duration time.Duration, audits []Audit) error {
	if botToken == "" || chatID == "" {
		return nil
	}

	var totalISRCs, totalResolved, totalAdded, totalDeleted, totalMissed, totalErrors int
	var detailsBuf bytes.Buffer
	var unchangedNames []string

	for _, a := range audits {
		totalISRCs += a.ISRCsFound
		totalResolved += a.TracksResolved
		totalAdded += a.TracksAdded
		totalDeleted += a.TracksDeleted
		totalMissed += a.TracksMissed

		if a.Err != nil {
			totalErrors++
		}

		isUnchanged := a.Err == nil && a.TracksAdded == 0 && a.TracksDeleted == 0 && a.TracksMissed == 0
		if isUnchanged {
			unchangedNames = append(unchangedNames, escapeMarkdown(a.Name))
			continue
		}

		statusText := "OK"
		if a.Err != nil {
			statusText = "FAILED"
		}

		fmt.Fprintf(&detailsBuf, "*» %s*\n", escapeMarkdown(a.Name))
		detailsBuf.WriteString("```\n")
		fmt.Fprintf(&detailsBuf, "Status:  %s\n", statusText)
		fmt.Fprintf(&detailsBuf, "ISRCs:   %d parsed | %d resolved", a.ISRCsFound, a.TracksResolved)
		if a.TracksMissed > 0 {
			fmt.Fprintf(&detailsBuf, " (%d missed)", a.TracksMissed)
		}
		detailsBuf.WriteString("\n")
		fmt.Fprintf(&detailsBuf, "Changes: +%d added | -%d deleted\n", a.TracksAdded, a.TracksDeleted)
		if a.Err != nil {
			errStr := strings.ReplaceAll(a.Err.Error(), "`", "'")
			fmt.Fprintf(&detailsBuf, "Error:   %s\n", errStr)
		}
		detailsBuf.WriteString("```\n\n")
	}

	var status string
	if totalErrors == 0 {
		status = "SUCCESS"
	} else if totalErrors == len(audits) {
		status = "FAILED"
	} else {
		status = "WARNING"
	}

	var unchangedSection string
	if len(unchangedNames) > 0 {
		if len(unchangedNames) == len(audits) {
			unchangedSection = "All playlists are up to date.\n"
		} else {
			unchangedSection = fmt.Sprintf("*Unchanged:* %s\n\n", strings.Join(unchangedNames, ", "))
		}
	}

	summaryText := fmt.Sprintf(
		"*DEEZER SYNC SUMMARY*\n"+
			"```\n"+
			"Status:    %s\n"+
			"Duration:  %s\n"+
			"Playlists: %d total (%d failed)\n"+
			"Tracks:    %d parsed | %d resolved\n"+
			"Changes:   +%d added | -%d deleted\n"+
			"```\n\n"+
			"%s"+
			"%s",
		status,
		duration.Round(time.Second),
		len(audits),
		totalErrors,
		totalISRCs,
		totalResolved,
		totalAdded,
		totalDeleted,
		unchangedSection,
		strings.TrimSpace(detailsBuf.String()),
	)

	payload := sendMessagePayload{
		ChatID:    chatID,
		Text:      strings.TrimSpace(summaryText),
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
