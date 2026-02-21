package tracker

import (
	"context"
	"errors"
	"testing"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/stretchr/testify/require"
)

func TestTorrentsPipelineStreamRequiresParser(t *testing.T) {
	tr := Init(Config{Urls: []string{"https://example.test"}})

	values, errs := tr.TorrentsPipelineStream(context.Background())

	_, ok := <-values
	require.False(t, ok)

	err, ok := <-errs
	require.True(t, ok)
	require.EqualError(t, err, "tracker parser is required")
}

func TestTorrentsPipelineStreamProcessesURLs(t *testing.T) {
	tr := Init(Config{
		Urls: []string{"u1", "u2"},
		TrackerParser: func(url string) ([]*torrents.Torrent, error) {
			if url == "" {
				return nil, errors.New("empty")
			}
			return []*torrents.Torrent{{Name: url}}, nil
		},
	})

	values, errs := tr.TorrentsPipelineStream(context.Background())

	var got []*torrents.Torrent
	for values != nil || errs != nil {
		select {
		case v, ok := <-values:
			if !ok {
				values = nil
				continue
			}
			got = append(got, v...)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			require.NoError(t, err)
		}
	}

	require.Len(t, got, 2)
}
