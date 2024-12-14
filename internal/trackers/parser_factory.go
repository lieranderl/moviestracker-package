package trackers

import "fmt"

type ParserFactory interface {
	CreateParser(config ParserConfig) (TorrentParser, error)
}

type DefaultParserFactory struct{}

func NewParserFactory() *DefaultParserFactory {
	return &DefaultParserFactory{}
}

func (f *DefaultParserFactory) CreateParser(config ParserConfig) (TorrentParser, error) {
	switch config.Type {
	case TrackerRutor:
		return NewRutorParser(config), nil
	case TrackerKinozal:
		return NewKinozalParser(config)
	default:
		return nil, fmt.Errorf("unsupported tracker type: %s", config.Type)
	}
}
