package isrchunt

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"musync/logger"
	"strconv"
	"time"
)

var isrcRegex *regexp.Regexp = regexp.MustCompile(`<td class="isrc">\s*([A-Z0-9]+)\s*</td>`)
var pageRegex *regexp.Regexp = regexp.MustCompile(`<(?:a|span)\s+class="page-link"[^>]*>(\d+)</(?:a|span)>`)

func ParseFullPlaylist(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	maxPage := 1
	for _, m := range pageRegex.FindAllStringSubmatch(html, -1) {
		page, _ := strconv.Atoi(m[1])
		if page > maxPage {
			maxPage = page
		}
	}
	logger.Infof("found %d pages", maxPage)

	seen := make(map[string]bool)
	var isrcIDs []string
	var dupesSkipped int

	addUnique := func(matches [][]string) int {
		added := 0
		for _, m := range matches {
			isrc := m[1]
			if seen[isrc] {
				dupesSkipped++
				continue
			}
			seen[isrc] = true
			isrcIDs = append(isrcIDs, isrc)
			added++
		}
		return added
	}

	page1Matches := isrcRegex.FindAllStringSubmatch(html, -1)
	added := addUnique(page1Matches)
	logger.Infof("parsed page 1/%d (%d found, %d unique, total=%d)",
		maxPage, len(page1Matches), added, len(isrcIDs))

	for page := 2; page <= maxPage; page++ {
		time.Sleep(200 * time.Millisecond) // basic politeness delay between page fetches

		resp, err := http.Get(fmt.Sprintf("%s&page=%d", url, page))
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		matches := isrcRegex.FindAllStringSubmatch(string(body), -1)
		added := addUnique(matches)

		logger.Infof("parsed page %d/%d (%d found, %d unique, total=%d)",
			page, maxPage, len(matches), added, len(isrcIDs))
	}

	logger.Infof("finished: %d total unique ISRCs (%d duplicates skipped)", len(isrcIDs), dupesSkipped)

	return isrcIDs, nil
}

func ParseToSync(url string) ([]string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var isrcIDs []string

	matches := isrcRegex.FindAllStringSubmatch(string(body), -1)

	for _, m := range matches {
		isrcIDs = append(isrcIDs, m[1])
	}

	return isrcIDs, nil
}
