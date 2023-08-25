package rutor

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"

	"github.com/lieranderl/moviestracker-package/internal/torrents"

	"github.com/gocolly/colly"
)

func ParseMoviePage(url string) ([]*torrents.Torrent, error) {
	pat := regexp.MustCompile(`btih:([aA-fF,0-9]{40})`)
	torrents := make([]*torrents.Torrent, 0)
	c := colly.NewCollector()
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		class := e.Attr("class")
		if class == "gai" || class == "tum" {
			if !strings.Contains(e.Text, "[") {
				t := new(rutorTorrent)
				t.rutorTitleToMovie(e.Text)
				t.Magnet, _ = e.DOM.Children().Eq(1).Children().Eq(1).Attr("href")
				t.MagnetHash = pat.FindAllStringSubmatch(t.Magnet, 1)[0][1]
				t.DetailsUrl, _ = e.DOM.Children().Eq(1).Children().Eq(2).Attr("href")
				t.DetailsUrl = "http://rutor.is" + t.DetailsUrl
				if t.OriginalName == "" {
					t.Hash = fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName+t.OriginalName+t.Year)))
				} else {
					t.Hash = fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName+t.Year)))
				}
				torrents = append(torrents, &t.Torrent)
			}
		}
	})
	err := c.Visit(url)

	return torrents, err
}

func ParseSeriesPage(url string) ([]*torrents.Torrent, error) {
	pat := regexp.MustCompile(`btih:([aA-fF,0-9]{40})`)
	torrents := make([]*torrents.Torrent, 0)
	c := colly.NewCollector()
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		class := e.Attr("class")
		if class == "gai" || class == "tum" {
			if strings.Contains(e.Text, "[") {
				t := new(rutorTorrent)
				t.rutorTitleToMovie(e.Text)
				t.Magnet, _ = e.DOM.Children().Eq(1).Children().Eq(1).Attr("href")
				t.MagnetHash = pat.FindAllStringSubmatch(t.Magnet, 1)[0][1]
				t.DetailsUrl, _ = e.DOM.Children().Eq(1).Children().Eq(2).Attr("href")
				t.DetailsUrl = "http://rutor.is" + t.DetailsUrl
				if t.OriginalName == "" {
					t.Hash = fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName+t.OriginalName+t.Year)))
				} else {
					t.Hash = fmt.Sprintf("%x", md5.Sum([]byte(t.RussianName+t.Year)))
				}
				torrents = append(torrents, &t.Torrent)
			}
		}
	})
	err := c.Visit(url)

	return torrents, err
}
