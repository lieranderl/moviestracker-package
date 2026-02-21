package kinozal

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Cred struct {
	Login    string
	Password string
}

const KINOZALLOGINURL = "https://kinozal.tv/takelogin.php"
const defaultRequestTimeout = 20 * time.Second

func Login(httpClient *http.Client, cred *Cred) (bool, *http.Client) {
	_, err := httpClient.Get(KINOZALLOGINURL)
	if err != nil {
		slog.Warn("cannot reach kinozal login", "error", err)
		return false, nil
	}

	time.Sleep(500 * time.Millisecond)
	_, err = httpClient.PostForm(KINOZALLOGINURL, url.Values{"username": {cred.Login}, "password": {cred.Password}, "wact": {"takerecover"}, "touser": {"1"}})
	if err != nil {
		slog.Warn("kinozal login attempt failed", "error", err)
		return false, nil
	}
	time.Sleep(500 * time.Millisecond)
	u, _ := url.Parse(KINOZALLOGINURL)
	for _, j := range httpClient.Jar.Cookies(u) {
		if j.Name == "pass" {
			return true, httpClient
		}
	}
	slog.Warn("kinozal login cookie not found")
	return false, nil
}

func GetMagnet(httpClient *http.Client, id string, mc chan map[string]string) {
	magnet, err := getMagnetForID(httpClient, id)
	if err != nil {
		slog.Warn("failed to resolve magnet", "details_id", id, "error", err)
	}
	mc <- map[string]string{id: magnet}
}

func getMagnetForID(httpClient *http.Client, id string) (string, error) {
	var magnet string
	bb, err := get(httpClient, "http://kinozal.tv/get_srv_details.php?id="+id+"&action=2")
	time.Sleep(300 * time.Millisecond)
	if err != nil {
		return "", err
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
	return magnet, nil
}

func get(httpClient *http.Client, url1 string) ([]byte, error) {
	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15")
	headers.Set("Accept-Encoding", "gzip")
	urlp, err := url.Parse(url1)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, urlp.String(), nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()
	req = req.WithContext(ctx)
	req.Header = headers

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check that the server actual sent compressed data
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}
	bb, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	} else {
		return bb, nil
	}
}
