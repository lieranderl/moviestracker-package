package trackers

import (
	"crypto/md5"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type RutorParser struct {
	BaseParser
	magnetPattern *regexp.Regexp
}

func NewRutorParser(config ParserConfig) *RutorParser {
	return &RutorParser{
		BaseParser:    BaseParser{config: config},
		magnetPattern: regexp.MustCompile(`btih:([aA-fF,0-9]{40})`),
	}
}

func (p *RutorParser) ParseMoviePage(url string) ([]*torrents.Torrent, error) {
	return p.parsePage(url, false)
}

func (p *RutorParser) ParseSeriesPage(url string) ([]*torrents.Torrent, error) {
	return p.parsePage(url, true)
}

func (p *RutorParser) parsePage(url string, isSeries bool) ([]*torrents.Torrent, error) {
	var results []*torrents.Torrent
	c := colly.NewCollector()

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		class := e.Attr("class")
		if class != "gai" && class != "tum" {
			return
		}

		isCurrent := false
		if isSeries {
			isCurrent = strings.Contains(e.Text, p.config.ParseConfig.SeriesPattern)
		} else {
			isCurrent = !strings.Contains(e.Text, p.config.ParseConfig.SeriesPattern)
		}

		if isCurrent {
			if torrent := p.parseTorrentElement(e); torrent != nil {
				results = append(results, torrent)
			}
		}
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to visit page: %w", err)
	}

	return results, nil
}

func (p *RutorParser) parseTorrentElement(e *colly.HTMLElement) *torrents.Torrent {
	torrent := &torrents.Torrent{}
	// Parse title and attributes
	listst := strings.Split(e.Text, "\n")
	if len(listst) < 3 {
		return nil
	}

	// Parse title
	title := p.parseTitle(listst[1])
	torrent.Name = title.Name
	torrent.RussianName = title.RussianName
	torrent.OriginalName = title.OriginalName
	torrent.Year = title.Year

	// Parse date
	if date, err := ParseDate(listst[0], p.config.ParseConfig.DateFormat); err == nil {
		torrent.Date = date
	}

	// Parse size and peers
	p.parseSizePeers(listst[2], torrent)

	// Parse magnet and details
	torrent.Magnet, _ = e.DOM.Children().Eq(1).Children().Eq(1).Attr("href")
	if matches := p.magnetPattern.FindStringSubmatch(torrent.Magnet); len(matches) > 1 {
		torrent.MagnetHash = matches[1]
	}

	detailsURL, _ := e.DOM.Children().Eq(1).Children().Eq(2).Attr("href")
	torrent.DetailsUrl = p.config.BaseURL + detailsURL

	// Generate hash
	torrent.Hash = p.generateHash(torrent)

	// Parse video attributes
	attrs := p.ParseVideoAttributes(title.Name)
	torrent.K4 = attrs.K4
	torrent.DV = attrs.DV
	torrent.FHD = attrs.FHD
	torrent.HDR = attrs.HDR
	torrent.HDR10 = attrs.HDR10
	torrent.HDR10plus = attrs.HDR10plus

	return torrent
}

type ParsedTitle struct {
	Name         string
	RussianName  string
	OriginalName string
	Year         string
}

func (p *RutorParser) parseTitle(text string) ParsedTitle {
	var title ParsedTitle

	if !strings.Contains(text, "/") {
		// Russian only title
		russianName, after, _ := strings.Cut(text, " (")
		title.Year = strings.Split(after, ")")[0]
		title.RussianName = russianName
	} else {
		// Title with original name
		parts := strings.Split(text, " / ")
		title.RussianName = parts[0]
		var originalName string
		originalName, after, _ := strings.Cut(parts[len(parts)-1], " (")
		title.Year = strings.Split(after, ")")[0]
		title.OriginalName = originalName
	}
	title.Name = text

	return title
}

func (p *RutorParser) parseSizePeers(text string, torrent *torrents.Torrent) {
	fields := strings.Fields(text)
	if len(fields) >= 4 {
		if size, err := ParseSize(fields[0], ""); err == nil {
			torrent.Size = size
		}
		if seeds, err := ParseInt(fields[2]); err == nil {
			torrent.Seeds = seeds
		}
		if leeches, err := ParseInt(fields[3]); err == nil {
			torrent.Leeches = leeches
		}
	}
}

func (p *RutorParser) generateHash(t *torrents.Torrent) string {
	if t.OriginalName == "" {
		return fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName)))
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName+t.OriginalName)))
}
