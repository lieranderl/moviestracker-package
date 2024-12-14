package trackers

import (
	"fmt"
	"strings"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type TrackerProcessor struct {
	parsers map[TrackerType]TorrentParser
	isMovie bool
}

func NewTrackerProcessor(isMovie bool, configs ...ParserConfig) (*TrackerProcessor, error) {
	factory := NewParserFactory()
	parsers := make(map[TrackerType]TorrentParser)

	for _, config := range configs {
		parser, err := factory.CreateParser(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create parser for %s: %w", config.Name, err)
		}
		parsers[config.Type] = parser
	}

	return &TrackerProcessor{
		parsers: parsers,
		isMovie: isMovie,
	}, nil
}

func (tp *TrackerProcessor) ProcessURL(url string) ([]*torrents.Torrent, error) {
	trackerType := determineTrackerType(url)
	parser, exists := tp.parsers[trackerType]
	
	if !exists {
		return nil, fmt.Errorf("no parser found for URL: %s", url)
	}

	if tp.isMovie {
		return parser.ParseMoviePage(url)
	}
	return parser.ParseSeriesPage(url)
}

func determineTrackerType(url string) TrackerType {
	switch {
	case strings.Contains(url, "rutor"):
		return TrackerRutor
	case strings.Contains(url, "kinozal"):
		return TrackerKinozal
	default:
		return TrackerType("unknown")
	}
}
