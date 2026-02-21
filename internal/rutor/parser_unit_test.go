package rutor

import (
	"testing"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/stretchr/testify/require"
)

func TestExtractMagnetHash(t *testing.T) {
	hash, ok := extractMagnetHash("magnet:?xt=urn:btih:5A64F87D68A7E4DF9F0E3F26B4E4C65C09D6FFAA")
	require.True(t, ok)
	require.Equal(t, "5A64F87D68A7E4DF9F0E3F26B4E4C65C09D6FFAA", hash)

	hash, ok = extractMagnetHash("not-a-magnet")
	require.False(t, ok)
	require.Empty(t, hash)
}

func TestBuildMovieHashStableForTrimmedAndCase(t *testing.T) {
	a := &rutorTorrent{
		Torrent: torrents.Torrent{
			RussianName:  " Плохие парни ",
			OriginalName: " Bad Boys ",
			Year:         "2024",
		},
	}
	b := &rutorTorrent{
		Torrent: torrents.Torrent{
			RussianName:  "плохие парни",
			OriginalName: "bad boys",
			Year:         "2024",
		},
	}

	require.Equal(t, buildMovieHash(a), buildMovieHash(b))
}
