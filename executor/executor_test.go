package executor

import (
	"errors"
	"strings"
	"testing"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/stretchr/testify/require"
)

func TestHandleErrorsReturnsCombinedError(t *testing.T) {
	pipeline := &TrackersPipeline{
		errors: []error{
			errors.New("first error"),
			nil,
			errors.New("second error"),
		},
	}

	err := pipeline.HandleErrors()
	require.Error(t, err)
	require.Contains(t, err.Error(), "first error")
	require.Contains(t, err.Error(), "second error")
}

func TestRunTrackersSearchPipelineValidatesURLs(t *testing.T) {
	pipeline := &TrackersPipeline{
		config: config{
			urls: []string{"https://rutor.example"},
		},
	}

	pipeline.RunTrackersSearchPipeline(true)
	err := pipeline.HandleErrors()

	require.Error(t, err)
	require.Contains(t, err.Error(), "at least two tracker urls")
}

func TestConvertTorrentsToMovieShortUsesStableGroupingKey(t *testing.T) {
	pipeline := &TrackersPipeline{
		torrents: []*torrents.Torrent{
			{
				Hash:         "hash-a",
				OriginalName: "Movie A",
				Year:         "2024",
				Date:         "2024-06-12T00:00:00.000Z",
			},
			{
				Hash:         "hash-a",
				OriginalName: "Movie A",
				Year:         "2024",
				Date:         "2024-07-12T00:00:00.000Z",
			},
			{
				MagnetHash:  "mh-b",
				RussianName: "Фильм Б",
				Year:        "2023",
				Date:        "2023-06-12T00:00:00.000Z",
			},
		},
	}

	pipeline.ConvertTorrentsToMovieShort()
	require.Len(t, pipeline.movies, 2)

	first := pipeline.movies[0]
	require.Equal(t, "Movie A", strings.TrimSpace(first.Searchname))
	require.Equal(t, "2024", first.Year)
	require.False(t, first.LastTimeFound.IsZero())

	second := pipeline.movies[1]
	require.Equal(t, "Фильм Б", strings.TrimSpace(second.Searchname))
	require.Equal(t, "2023", second.Year)
}
