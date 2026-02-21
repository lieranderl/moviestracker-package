package tracker

import (
	"context"
	"errors"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/lieranderl/moviestracker-package/pkg/pipeline"
)

type Config struct {
	Urls          []string
	TrackerParser func(string) ([]*torrents.Torrent, error)
}

type Tracker struct {
	urls          []string
	trackerParser func(string) ([]*torrents.Torrent, error)
}

func Init(config Config) *Tracker {
	return &Tracker{
		urls:          append([]string(nil), config.Urls...),
		trackerParser: config.TrackerParser,
	}
}

func (t Tracker) TorrentsPipelineStream(ctx context.Context) (<-chan []*torrents.Torrent, <-chan error) {
	if t.trackerParser == nil {
		tc := make(chan []*torrents.Torrent)
		ec := make(chan error, 1)
		close(tc)
		ec <- errors.New("tracker parser is required")
		close(ec)
		return tc, ec
	}

	urlStream, err := pipeline.Producer(ctx, t.urls)
	if err != nil {
		tc := make(chan []*torrents.Torrent)
		ec := make(chan error, 1)
		close(tc)
		ec <- err
		close(ec)
		return tc, ec
	}
	torrents_chan, errors := pipeline.Step(ctx, urlStream, t.trackerParser, 3)
	return torrents_chan, errors
}
