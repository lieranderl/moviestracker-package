// common.go
package trackers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

// BaseParser provides common parsing functionality
type BaseParser struct {
	config ParserConfig
}

// ParseVideoAttributes implements common video attribute parsing
func (p *BaseParser) ParseVideoAttributes(text string) *torrents.Torrent {
	t := &torrents.Torrent{}
	textLower := strings.ToLower(text)

	if strings.Contains(textLower, "2160p") {
		t.K4 = true
	}
	if strings.Contains(textLower, "dolby vision") {
		t.DV = true
	}
	if strings.Contains(text, "1080p") {
		t.FHD = true
	}
	if strings.Contains(textLower, "hdr10+") {
		t.HDR10plus = true
		t.HDR10 = true
		t.HDR = true
	} else if strings.Contains(textLower, "hdr10") {
		t.HDR10 = true
		t.HDR = true
	} else if strings.Contains(textLower, "hdr") {
		t.HDR = true
	}

	return t
}

// ParseSize parses size string to float32
func ParseSize(s string, suffix string) (float32, error) {
	value := strings.TrimSpace(strings.TrimSuffix(s, suffix))
	if f, err := strconv.ParseFloat(value, 32); err == nil {
		return float32(f), nil
	}
	return 0, fmt.Errorf("failed to parse size")
}

// ParseInt parses string to int32
func ParseInt(s string) (int32, error) {
	value := strings.TrimSpace(s)
	if i, err := strconv.ParseInt(value, 10, 32); err == nil {
		return int32(i), nil
	}
	return 0, fmt.Errorf("failed to parse int")
}

// ParseDate parses date string using specified format
func ParseDate(dateStr string, format string) (string, error) {
	// ISO 8601 format
	layout := "2006-01-02"
	timeNow := time.Now()
	switch {
	case strings.Contains(dateStr, "сегодня"):
		// timeNow to 2006-01-02 format
		return timeNow.Format(layout), nil
	case strings.Contains(dateStr, "вчера"):
		return timeNow.AddDate(0, 0, -1).Format(layout), nil
	default:
		cleanDateStr := strings.ReplaceAll(dateStr, "\u00A0", " ")
		if parsedTime, err := monday.Parse(format, cleanDateStr, monday.LocaleRuRU); err == nil {
			return parsedTime.Format(layout), nil
		}
		return "", fmt.Errorf("failed to parse date: %s", dateStr)
	}
}
