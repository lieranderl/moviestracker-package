package trackers

import "time"

type ParserConfig struct {
	Type          TrackerType
	Name          string
	BaseURL       string
	RequiresAuth  bool
	AuthConfig    *AuthConfig    `json:",omitempty"`
	ParseConfig   *ParseConfig   `json:",omitempty"`
	RequestConfig *RequestConfig `json:",omitempty"`
}

type AuthConfig struct {
	LoginURL       string
	LoginEnvVar    string
	PasswordEnvVar string
}

type ParseConfig struct {
	MoviePattern    string
	SeriesPattern   string
	ExcludePatterns []string
	DateFormat      string
	TitleSelector   string
	SizeSelector    string
	SeedsSelector   string
	LeechSelector   string
}

type RequestConfig struct {
	Timeout        time.Duration
	RetryCount     int
	RetryDelay     time.Duration
	UserAgent      string
	AcceptLanguage string
}

var DefaultRequestConfig = &RequestConfig{
	Timeout:        30 * time.Second,
	RetryCount:     3,
	RetryDelay:     time.Second,
	UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	AcceptLanguage: "en-US,en;q=0.9",
}
