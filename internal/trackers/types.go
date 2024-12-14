package trackers

import (
	"context"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type TrackerType string

const (
	TrackerRutor   TrackerType = "rutor"
	TrackerKinozal TrackerType = "kinozal"
)

// TorrentParser defines interface for parsing torrents
type TorrentParser interface {
	ParseMoviePage(url string) ([]*torrents.Torrent, error)
	ParseSeriesPage(url string) ([]*torrents.Torrent, error)
}

// ResultMerger defines how to merge results from multiple sources
type ResultMerger interface {
	MergeTorrents(ctx context.Context, channels ...chan []*torrents.Torrent) chan []*torrents.Torrent
	MergeErrors(ctx context.Context, channels ...chan error) chan error
}

// SearchResult represents search operation outcome
type SearchResult struct {
	Torrents []*torrents.Torrent
	Errors   []error
}
