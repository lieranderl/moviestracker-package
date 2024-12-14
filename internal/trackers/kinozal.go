// kinozal.go
package trackers

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type KinozalParser struct {
	BaseParser
	client     *http.Client
	isLoggedIn bool
}

func NewKinozalParser(config ParserConfig) (*KinozalParser, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: config.RequestConfig.Timeout,
	}

	return &KinozalParser{
		BaseParser: BaseParser{config: config},
		client:     client,
	}, nil
}

func (p *KinozalParser) login() error {
	if p.isLoggedIn {
		return nil
	}

	username := os.Getenv(p.config.AuthConfig.LoginEnvVar)
	password := os.Getenv(p.config.AuthConfig.PasswordEnvVar)
	log.Println("login url", p.config.AuthConfig.LoginURL)

	if username == "" || password == "" {
		return fmt.Errorf("missing credentials in environment variables")
	}

	formData := url.Values{
		"username": {username},
		"password": {password},
		"wact":     {"takerecover"},
		"touser":   {"1"},
	}

	req, err := http.NewRequest("POST", p.config.AuthConfig.LoginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", p.config.RequestConfig.UserAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	resp, err := p.client.Do(req)
	if err != nil {
		log.Fatalln("Login attempt failed")
	}

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (p *KinozalParser) verifyLogin() bool {
	u, err := url.Parse(p.config.AuthConfig.LoginURL)
	if err != nil {
		return false
	}

	for _, cookie := range p.client.Jar.Cookies(u) {
		log.Println(cookie.Name)
		if cookie.Name == "pass" {
			return true
		}
	}
	log.Println("No Login")
	return false
}

func (p *KinozalParser) ParseMoviePage(url string) ([]*torrents.Torrent, error) {
	if err := p.login(); err != nil {
		return nil, err
	}
	return p.parsePage(url, false)
}

func (p *KinozalParser) ParseSeriesPage(url string) ([]*torrents.Torrent, error) {
	if err := p.login(); err != nil {
		return nil, err
	}
	return p.parsePage(url, true)
}

func (p *KinozalParser) parsePage(url string, isSeries bool) ([]*torrents.Torrent, error) {
	httpcl := p.client
	c := colly.NewCollector()
	var results []*torrents.Torrent

	c.OnHTML(p.config.ParseConfig.TitleSelector, func(e *colly.HTMLElement) {
		name := e.DOM.Children().Eq(1).Text()

		isCurrent := false
		if isSeries {
			isCurrent = strings.Contains(name, p.config.ParseConfig.SeriesPattern)
		} else {
			isCurrent = !strings.Contains(name, p.config.ParseConfig.SeriesPattern)
		}

		if isCurrent {
			if torrent := p.parseTorrentElement(e.DOM.Children(), name); torrent != nil {
				results = append(results, torrent)
			}
		}
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to visit page: %w", err)
	}

	if len(results) > 0 && p.verifyLogin() {
		results = fetchMagnetLinks(results, httpcl)
	}

	return results, nil
}

func (p *KinozalParser) parseTorrentElement(tds *goquery.Selection, name string) *torrents.Torrent {
	torrent := p.ParseVideoAttributes(name)
	torrent.Name = name

	if size, err := ParseSize(tds.Eq(3).Text(), " ГБ"); err == nil {
		torrent.Size = size
	}
	if seeds, err := ParseInt(tds.Eq(4).Text()); err == nil {
		torrent.Seeds = seeds
	}
	if leeches, err := ParseInt(tds.Eq(5).Text()); err == nil {
		torrent.Leeches = leeches
	}
	if dateStr := strings.Fields(tds.Eq(6).Text()); len(dateStr) > 0 {
		if date, err := ParseDate(dateStr[0], p.config.ParseConfig.DateFormat); err == nil {
			torrent.Date = date
		}
	}

	torrent.DetailsUrl = strings.TrimSpace(strings.Split(tds.Find("a").First().AttrOr("href", ""), "id=")[1])
	return torrent
}

func fetchMagnetLinks(torrents []*torrents.Torrent, httpcl *http.Client) []*torrents.Torrent {
	pat := regexp.MustCompile(`btih:([aA-fF,0-9]{40})`)
	magnetChannel := make(chan map[string]string)
	for _, m := range torrents {
		go GetMagnet(httpcl, m.DetailsUrl, magnetChannel)
	}
	i := 0
	for mc := range magnetChannel {
		i += 1
		for _, m := range torrents {
			if id, ok := mc[m.DetailsUrl]; ok {
				m.Magnet = id
				m.MagnetHash = pat.FindAllStringSubmatch(id, 1)[0][1]
				break
			}
		}
		if i == len(torrents) {
			close(magnetChannel)
		}
	}
	return torrents
}

func GetMagnet(httpClient *http.Client, id string, mc chan map[string]string) {
	var magnet string
	bb, err := get(httpClient, "https://kinozal.tv/get_srv_details.php?id="+id+"&action=2")
	time.Sleep(1 * time.Second)
	if err != nil {
		log.Fatalln("failed to get data.")
	}
	if strings.Contains(string(bb), "хеш:") {
		list := strings.Fields(string(bb))
		for i, s := range list {
			if strings.Contains(string(s), "хеш:") {
				magnet = "magnet:?xt=urn:btih:" + strings.Split(list[i+1], "<")[0]
				break
			}
		}
	}
	mc <- map[string]string{id: magnet}
}

func get(httpClient *http.Client, url1 string) ([]byte, error) {
	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15")
	headers.Set("Accept-Encoding", "gzip")
	urlp, err := url.Parse(url1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req := &http.Request{Header: headers, URL: urlp}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check that the server actual sent compressed data
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Println("Gzip error", err)
			return nil, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}
	bb, err := io.ReadAll(reader)
	if err != nil {
		log.Println("Read error", err)
		return nil, err
	} else {
		return bb, nil
	}
}
