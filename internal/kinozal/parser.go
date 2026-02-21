package kinozal

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/goodsign/monday"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

var (
	kzBTIHPattern = regexp.MustCompile(`btih:([a-fA-F0-9]{40})`)
	kzYearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)
)

func kzLogin() (bool, *http.Client) {
	cred := &Cred{
		Login:    os.Getenv("KZ_LOGIN"),
		Password: os.Getenv("KZ_PASSWORD"),
	}

	if cred.Login == "" || cred.Password == "" {
		return false, nil
	}

	jar, _ := cookiejar.New(&cookiejar.Options{})
	httpClient := &http.Client{
		Jar:       jar,
		Transport: &http.Transport{},
		Timeout:   defaultRequestTimeout,
	}

	return Login(httpClient, cred)
}

func parseDetailsID(href string) string {
	parts := strings.Split(strings.TrimSpace(href), "id=")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

func ParseMoviePage(url string) ([]*torrents.Torrent, error) {
	return parsePage(url, false)
}

func ParseSeriesPage(url string) ([]*torrents.Torrent, error) {
	return parsePage(url, true)
}

func parsePage(url string, isSeries bool) ([]*torrents.Torrent, error) {
	loggedIn, httpClient := kzLogin()
	titles := make([]*torrents.Torrent, 0)

	c := colly.NewCollector()
	c.SetRequestTimeout(defaultRequestTimeout)
	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15"
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "ru,en;q=0.8")
	})
	c.OnHTML("tr.bg", func(e *colly.HTMLElement) {
		t := new(kzTorrent)
		t.DetailsUrl = parseDetailsID(e.ChildAttr("a", "href"))
		if t.DetailsUrl == "" {
			return
		}

		tds := e.DOM.Children()
		t.Name = strings.TrimSpace(tds.Eq(1).Text())
		containsSeriesMarker := strings.Contains(strings.ToLower(t.Name), "сезон")
		if containsSeriesMarker != isSeries {
			return
		}

		t.parseHtmlTor(tds)
		titles = append(titles, &t.Torrent)
	})

	err := c.Visit(url)
	if err != nil {
		return titles, fmt.Errorf("kinozal visit %q: %w", url, err)
	}

	if loggedIn && len(titles) > 0 {
		titles = fetchMagnetLinks(titles, httpClient)
	}

	return titles, nil
}

func containsCyrillic(s string) bool {
	for _, r := range s {
		if (r >= 'А' && r <= 'я') || r == 'Ё' || r == 'ё' {
			return true
		}
	}
	return false
}

func (m *kzTorrent) parseTitleMetadata() {
	name := strings.Join(strings.Fields(strings.TrimSpace(m.Name)), " ")
	if name == "" {
		return
	}

	if year := kzYearPattern.FindString(name); year != "" {
		m.Year = year
	}

	baseName := name
	if m.Year != "" {
		if idx := strings.Index(baseName, m.Year); idx > 0 {
			baseName = strings.TrimSpace(baseName[:idx])
		}
	}
	baseName = strings.Trim(baseName, "[]()|/- ")

	parts := strings.Split(baseName, " / ")
	switch {
	case len(parts) >= 2:
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[len(parts)-1])

		switch {
		case containsCyrillic(left) && !containsCyrillic(right):
			m.RussianName = left
			m.OriginalName = right
		case !containsCyrillic(left) && containsCyrillic(right):
			m.OriginalName = left
			m.RussianName = right
		default:
			m.RussianName = left
			m.OriginalName = right
		}
	default:
		if containsCyrillic(baseName) {
			m.RussianName = baseName
		} else {
			m.OriginalName = baseName
		}
	}

	if m.RussianName == "" {
		m.RussianName = m.OriginalName
	}
	if m.OriginalName == "" {
		m.OriginalName = m.RussianName
	}

	hashSource := strings.ToLower(strings.Join([]string{
		strings.TrimSpace(m.RussianName),
		strings.TrimSpace(m.OriginalName),
		strings.TrimSpace(m.Year),
	}, "|"))
	m.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(hashSource)))
}

func (m *kzTorrent) parseHtmlTor(tds *goquery.Selection) {
	m.parseAttributes(m.Name)
	m.parseTitleMetadata()

	if s, err := strconv.ParseFloat(strings.Split(tds.Eq(3).Text(), " ГБ")[0], 32); err == nil {
		m.Size = float32(s)
	}
	if s, err := strconv.ParseInt(strings.TrimSpace(tds.Eq(4).Text()), 10, 32); err == nil {
		m.Seeds = int32(s)
	}
	if s, err := strconv.ParseInt(strings.TrimSpace(tds.Eq(5).Text()), 10, 32); err == nil {
		m.Leeches = int32(s)
	}

	const (
		layout       = "2006-01-02T00:00:00.000Z"
		layoutParsed = "02.01.2006"
	)

	timeNow := time.Now()
	yesterday := time.Now().Add(-24 * time.Hour)

	if len(strings.Fields(tds.Eq(6).Text())) > 0 {
		d := strings.Fields(tds.Eq(6).Text())[0]
		switch {
		case strings.Contains(d, "сегодня"):
			m.Date = monday.Format(timeNow, layout, monday.LocaleRuRU)
		case strings.Contains(d, "вчера"):
			m.Date = monday.Format(yesterday, layout, monday.LocaleRuRU)
		default:
			parsedMtime, err := monday.Parse(layoutParsed, d, monday.LocaleRuRU)
			if err == nil {
				m.Date = monday.Format(parsedMtime, layout, monday.LocaleRuRU)
			}
		}
	}

	if m.Date == "" {
		m.Date = monday.Format(timeNow, layout, monday.LocaleRuRU)
	}
}

func fetchMagnetLinks(titles []*torrents.Torrent, httpClient *http.Client) []*torrents.Torrent {
	type magnetResult struct {
		detailsID  string
		magnetLink string
	}

	byDetailsID := make(map[string]*torrents.Torrent, len(titles))
	for _, movie := range titles {
		if movie == nil || movie.DetailsUrl == "" {
			continue
		}
		byDetailsID[movie.DetailsUrl] = movie
	}

	magnetChannel := make(chan magnetResult, len(byDetailsID))
	requests := 0
	for _, movie := range titles {
		if movie == nil || movie.DetailsUrl == "" {
			continue
		}
		requests++
		go func(detailsID string) {
			magnetLink, err := getMagnetForID(httpClient, detailsID)
			if err != nil {
				slog.Warn("failed to fetch kinozal magnet", "details_id", detailsID, "error", err)
			}
			magnetChannel <- magnetResult{
				detailsID:  detailsID,
				magnetLink: magnetLink,
			}
		}(movie.DetailsUrl)
	}

	for i := 0; i < requests; i++ {
		res := <-magnetChannel
		movie, ok := byDetailsID[res.detailsID]
		if !ok {
			continue
		}
		movie.Magnet = res.magnetLink
		matches := kzBTIHPattern.FindStringSubmatch(res.magnetLink)
		if len(matches) == 2 {
			movie.MagnetHash = matches[1]
		}
	}

	close(magnetChannel)
	return titles
}

type kzTorrent struct {
	torrents.Torrent
}

func (t *kzTorrent) parseAttributes(after string) {
	lower := strings.ToLower(after)

	if strings.Contains(lower, "2160p") {
		t.K4 = true
	}
	if strings.Contains(lower, "dolby vision") {
		t.DV = true
	}
	if strings.Contains(after, "1080p") {
		t.FHD = true
	}

	switch {
	case strings.Contains(lower, "hdr10+"):
		t.HDR10plus = true
		t.HDR10 = true
		t.HDR = true
	case strings.Contains(lower, "hdr10"):
		t.HDR10 = true
		t.HDR = true
	case strings.Contains(lower, "hdr"):
		t.HDR = true
	}
}
