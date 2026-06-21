package deezer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"musync/logger"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type deezerTrack struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type deezerErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

type cacheEntry struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func GetDeezerTrackIDsByISRCs(isrcs []string) ([]int64, error) {
	if len(isrcs) == 0 {
		return nil, nil
	}

	trackIDs := make([]int64, len(isrcs))
	cache := loadISRCCache()
	var missingIndices []int
	var missingISRCs []string

	for i, isrc := range isrcs {
		if entry, found := cache[isrc]; found && entry.ID != 0 {
			trackIDs[i] = entry.ID
		} else {
			missingIndices = append(missingIndices, i)
			missingISRCs = append(missingISRCs, isrc)
		}
	}

	if len(missingISRCs) == 0 {
		return trackIDs, nil
	}

	logger.Infof("resolving %d track IDs via Deezer API (%d cached)", len(missingISRCs), len(isrcs)-len(missingISRCs))

	total := len(missingISRCs)
	limiter := rate.NewLimiter(rate.Limit(9), 1)

	wg := sync.WaitGroup{}
	errCh := make(chan error, total)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	ctx := context.Background()

	var completed atomic.Int64

	tickerDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				done := completed.Load()
				logger.Info("isrc progress", "done", done, "remaining", int64(total)-done, "total", total)
			case <-tickerDone:
				return
			}
		}
	}()

	type resolvedResult struct {
		isrc  string
		id    int64
		title string
	}
	resultsCh := make(chan resolvedResult, total)

	for i, isrc := range missingISRCs {
		wg.Add(1)
		origIdx := missingIndices[i]
		go func(origIdx int, isrc string) {
			defer wg.Done()
			defer func() {
				completed.Add(1)
			}()

			if err := limiter.Wait(ctx); err != nil {
				errCh <- err
				return
			}

			url := fmt.Sprintf("https://api.deezer.com/track/isrc:%s", isrc)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				errCh <- err
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0")

			resp, err := client.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errCh <- err
				return
			}

			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("isrc %s: unexpected status %s", isrc, resp.Status)
				return
			}

			var maybeErr deezerErrorResponse
			if err := json.Unmarshal(body, &maybeErr); err == nil && maybeErr.Error.Type != "" {
				errCh <- fmt.Errorf("isrc %s: deezer error [%s, code %d]: %s",
					isrc, maybeErr.Error.Type, maybeErr.Error.Code, maybeErr.Error.Message)
				return
			}

			var t deezerTrack
			if err := json.Unmarshal(body, &t); err != nil {
				errCh <- fmt.Errorf("isrc %s: unmarshal failed: %w", isrc, err)
				return
			}
			if t.ID == 0 {
				errCh <- fmt.Errorf("no track found for ISRC: %s", isrc)
				return
			}

			trackIDs[origIdx] = t.ID
			resultsCh <- resolvedResult{isrc: isrc, id: t.ID, title: t.Title}
		}(origIdx, isrc)
	}
	wg.Wait()
	close(tickerDone)
	close(errCh)
	close(resultsCh)

	var failCount int
	for err := range errCh {
		failCount++
		logger.Info("isrc lookup error", "err", err)
	}

	logger.Info("=== ISRC RESOLUTION DONE ===",
		"resolved", total-failCount,
		"total", total,
		"failed", failCount,
	)

	newlyResolved := 0
	for res := range resultsCh {
		cache[res.isrc] = cacheEntry{
			ID:    res.id,
			Title: res.title,
		}
		newlyResolved++
	}
	if newlyResolved > 0 {
		saveISRCCache(cache)
	}

	return trackIDs, nil
}

func loadISRCCache() map[string]cacheEntry {
	cache := make(map[string]cacheEntry)
	data, err := os.ReadFile("isrc_cache.json")
	if err != nil {
		return cache
	}

	// Try loading as new format first
	if err := json.Unmarshal(data, &cache); err == nil {
		hasNewFormat := false
		for _, entry := range cache {
			if entry.ID != 0 {
				hasNewFormat = true
				break
			}
		}
		if hasNewFormat || len(cache) == 0 {
			return cache
		}
	}

	// If parsing failed or has no valid IDs, try migrating old map[string]int64 format
	oldCache := make(map[string]int64)
	if err := json.Unmarshal(data, &oldCache); err == nil {
		for isrc, id := range oldCache {
			cache[isrc] = cacheEntry{
				ID:    id,
				Title: "", // Title is unknown for migrated entries
			}
		}
		saveISRCCache(cache)
	}
	return cache
}

func saveISRCCache(cache map[string]cacheEntry) {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		logger.Error("failed to marshal ISRC cache", "err", err)
		return
	}
	if err := os.WriteFile("isrc_cache.json", data, 0644); err != nil {
		logger.Error("failed to save ISRC cache", "err", err)
	}
}
