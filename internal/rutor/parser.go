package rutor

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

var btihPattern = regexp.MustCompile(`btih:([a-fA-F0-9]{40})`)

func extractMagnetHash(magnet string) (string, bool) {
	matches := btihPattern.FindStringSubmatch(magnet)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func buildMovieHash(t *rutorTorrent) string {
	key := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(t.RussianName)),
		strings.ToLower(strings.TrimSpace(t.OriginalName)),
		strings.TrimSpace(t.Year),
	}, "|")
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

func parsePage(url string, isSeries bool) ([]*torrents.Torrent, error) {
	result := make([]*torrents.Torrent, 0)
	c := colly.NewCollector()
	c.SetRequestTimeout(20 * time.Second)
	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15"

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		class := e.Attr("class")
		if class != "gai" && class != "tum" {
			return
		}

		containsSeriesMarker := strings.Contains(e.Text, "[")
		if isSeries != containsSeriesMarker {
			return
		}

		t := new(rutorTorrent)
		t.rutorTitleToMovie(e.Text)

		magnet, exists := e.DOM.Children().Eq(1).Children().Eq(1).Attr("href")
		if !exists {
			return
		}
		t.Magnet = magnet

		magnetHash, ok := extractMagnetHash(magnet)
		if !ok {
			return
		}
		t.MagnetHash = magnetHash

		detailsURL, exists := e.DOM.Children().Eq(1).Children().Eq(2).Attr("href")
		if !exists {
			return
		}
		t.DetailsUrl = "https://rutor.is" + detailsURL
		t.Hash = buildMovieHash(t)

		result = append(result, &t.Torrent)
	})

	err := c.Visit(url)
	if err != nil {
		return result, fmt.Errorf("rutor visit %q: %w", url, err)
	}

	return result, nil
}

func ParseMoviePage(url string) ([]*torrents.Torrent, error) {
	return parsePage(url, false)
}

func ParseSeriesPage(url string) ([]*torrents.Torrent, error) {
	return parsePage(url, true)
}
