package main

import (
	"fmt"
	"musync/deezer"
	isrchunt "musync/isrcHunt"
	"musync/logger"
	"musync/telegram"
	"net/http"
	"os"
	"time"
)

func runSync(client *http.Client, deezerArl string, playlists []Playlist) {
	runStart := time.Now()
	logger.Info("[START]")

	deezerToken, deezerUserID, err := deezer.GetDeezerCreds(client, deezerArl)
	if err != nil {
		logger.Error("deezer auth failed", "err", err)
		return
	}

	session := &deezer.Session{
		HTTPClient: client,
		ARL:        deezerArl,
		Token:      deezerToken,
		UserID:     deezerUserID,
	}

	var audits []telegram.Audit

	for i := range playlists {
		audit, err := syncPlaylist(session, &playlists[i])
		if err != nil {
			logger.Error("failed to sync playlist", "playlist", playlists[i].Name, "err", err)
			audit.Err = err
		}
		audits = append(audits, audit)
	}

	if err := SavePlaylists(playlists); err != nil {
		logger.Error("failed to save playlists", "err", err)
	}

	printAndNotifySummary(audits, runStart)
}

func syncPlaylist(session *deezer.Session, playlist *Playlist) (telegram.Audit, error) {
	audit := telegram.Audit{Name: playlist.Name}

	ISRCs, err := isrchunt.ParseFullPlaylist(playlist.Url)
	if err != nil {
		return audit, fmt.Errorf("parse playlist: %w", err)
	}
	audit.ISRCsFound = len(ISRCs)
	logger.Infof("finished sync: %d total ISRCs", len(ISRCs))

	deezerTrackIDs, err := deezer.GetDeezerTrackIDsByISRCs(ISRCs)
	if err != nil {
		return audit, fmt.Errorf("get track IDs: %w", err)
	}

	var resolved []int64
	for _, id := range deezerTrackIDs {
		if id != 0 {
			resolved = append(resolved, id)
		}
	}
	audit.TracksResolved = len(resolved)
	audit.TracksMissed = len(ISRCs) - len(resolved)
	logger.Infof("[TRACK IDs FOUND]: %d / %d", len(resolved), len(ISRCs))

	deezerPlaylistID := playlist.DeezerID
	if deezerPlaylistID == 0 {
		deezerPlaylistID, err = deezer.GetOrCreatePlaylist(session, playlist.Name)
		if err != nil {
			return audit, fmt.Errorf("get or create playlist: %w", err)
		}
		playlist.updateDeezerID(deezerPlaylistID)
	}

	if len(resolved) == 0 {
		logger.Info("no tracks to add, skipping", "playlist", playlist.Name)
		return audit, nil
	}

	existing, err := deezer.GetPlaylistTrackIDs(session, deezerPlaylistID)
	if err != nil {
		return audit, fmt.Errorf("fetch existing tracks: %w", err)
	}

	var newTrackIDs []int64
	for _, id := range resolved {
		if existing[id] {
			audit.AlreadyPresent++
		} else {
			newTrackIDs = append(newTrackIDs, id)
		}
	}
	resolvedSet := make(map[int64]bool, len(resolved))
	for _, id := range resolved {
		resolvedSet[id] = true
	}

	var removeTrackIDs []int64
	for id := range existing {
		if !resolvedSet[id] {
			removeTrackIDs = append(removeTrackIDs, id)
		}
	}

	if len(newTrackIDs) == 0 && len(removeTrackIDs) == 0 {
		logger.Info("playlist already up to date")
		return audit, nil
	}
	if len(removeTrackIDs) > 0 {
		if err := deezer.DeleteSongsFromPlaylist(session, deezerPlaylistID, removeTrackIDs); err != nil {
			return audit, fmt.Errorf("remove songs: %w", err)
		}
		audit.TracksDeleted = len(removeTrackIDs)
	}

	if len(newTrackIDs) > 0 {
		if err := deezer.AddSongsToPlaylist(session, deezerPlaylistID, newTrackIDs); err != nil {
			return audit, fmt.Errorf("add songs: %w", err)
		}
		audit.TracksAdded = len(newTrackIDs)
	}

	logger.Info("sync completed", "playlist", playlist.Name, "added", len(newTrackIDs), "already_present", audit.AlreadyPresent, "deleted", len(removeTrackIDs))
	return audit, nil
}

func printAndNotifySummary(audits []telegram.Audit, runStart time.Time) {
	var totalISRCs, totalResolved, totalMissed, totalAlreadyPresent, totalAdded, totalErrors, totalDeleted int
	for _, a := range audits {
		totalISRCs += a.ISRCsFound
		totalResolved += a.TracksResolved
		totalMissed += a.TracksMissed
		totalAlreadyPresent += a.AlreadyPresent
		totalAdded += a.TracksAdded
		totalDeleted += a.TracksDeleted
		if a.Err != nil {
			totalErrors++
		}
	}

	logger.Info("=== SYNC RUN SUMMARY ===",
		"duration", time.Since(runStart).Round(time.Second),
		"playlists_processed", len(audits),
		"playlists_failed", totalErrors,
		"isrcs_found", totalISRCs,
		"tracks_resolved", totalResolved,
		"tracks_missed", totalMissed,
		"already_present", totalAlreadyPresent,
		"tracks_added", totalAdded,
		"tracks_deleted", totalDeleted,
	)

	for _, a := range audits {
		status := "ok"
		if a.Err != nil {
			status = "FAILED: " + a.Err.Error()
		}
		logger.Info("  playlist audit",
			"name", a.Name,
			"isrcs", a.ISRCsFound,
			"resolved", a.TracksResolved,
			"missed", a.TracksMissed,
			"already_present", a.AlreadyPresent,
			"added", a.TracksAdded,
			"deleted", a.TracksDeleted,
			"status", status,
		)
	}

	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgChatID := os.Getenv("TELEGRAM_CHAT_ID")
	if tgToken != "" && tgChatID != "" {
		if err := telegram.SendSyncSummary(tgToken, tgChatID, time.Since(runStart), audits); err != nil {
			logger.Error("failed to send telegram notification", "err", err)
		} else {
			logger.Info("telegram notification sent successfully")
		}
	}

	logger.Info("[DONE]")
}
