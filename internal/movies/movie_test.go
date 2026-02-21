package movies

import (
	"testing"
	"time"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/stretchr/testify/require"
)

func TestUpdateMoviesAttribsUsesLatestDate(t *testing.T) {
	m := &Short{
		Torrents: []*torrents.Torrent{
			{Date: "2024-01-01T00:00:00.000Z"},
			{Date: "2024-02-01T00:00:00.000Z"},
		},
	}

	m.UpdateMoviesAttribs()

	require.Equal(t, "2024-02-01T00:00:00.000Z", m.LastTimeFound.UTC().Format("2006-01-02T15:04:05.000Z"))
}

func TestSetLastTimeFoundFallbackOnInvalidDate(t *testing.T) {
	m := &Short{}
	before := time.Now().UTC().Add(-2 * time.Second)

	m.setLastTimeFound(&torrents.Torrent{Date: "invalid"})

	require.True(t, m.LastTimeFound.After(before))
}
