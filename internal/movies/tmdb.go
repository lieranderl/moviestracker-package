package movies

import (
	"context"
	"fmt"
	"strings"

	// "math/rand"

	"github.com/lieranderl/go-tmdb"
	"github.com/lieranderl/moviestracker-package/pkg/pipeline"
)

type TMDb struct {
	tmdb *tmdb.TMDb
}

func TMDBInit(tmdbkey string) *TMDb {
	var TMDBCONFIG = tmdb.Config{
		APIKey:   tmdbkey,
		Proxies:  nil,
		UseProxy: false,
	}
	mytmdb := new(TMDb)
	mytmdb.tmdb = tmdb.Init(TMDBCONFIG)
	return mytmdb
}

func (tmdbapi *TMDb) fetchMovieDetails(m *Short) (*Short, error) {
	options := make(map[string]string)
	options["language"] = "ru"
	options["year"] = m.Year
	r, err := tmdbapi.tmdb.SearchMovie(m.Searchname, options)
	if err != nil {
		return nil, fmt.Errorf("tmdb search %q (%s): %w", m.Searchname, m.Year, err)
	}
	if len(r.Results) > 0 {
		releaseDate := strings.TrimSpace(r.Results[0].ReleaseDate)
		releaseYear := ""
		if len(releaseDate) >= 4 {
			releaseYear = releaseDate[:4]
		}

		sameTitle := strings.EqualFold(strings.TrimSpace(m.Searchname), strings.TrimSpace(r.Results[0].OriginalTitle)) ||
			strings.EqualFold(strings.TrimSpace(m.Searchname), strings.TrimSpace(r.Results[0].Title))
		sameYear := m.Year == "" || releaseYear == "" || m.Year == releaseYear

		if sameTitle && sameYear {
			// m.Adult = r.Results[0].Adult
			m.BackdropPath = r.Results[0].BackdropPath
			m.ID = fmt.Sprint(r.Results[0].ID)
			m.OriginalTitle = r.Results[0].OriginalTitle
			m.GenreIDs = r.Results[0].GenreIDs
			// m.Popularity = r.Results[0].Popularity
			m.PosterPath = r.Results[0].PosterPath
			m.ReleaseDate = r.Results[0].ReleaseDate
			m.Title = r.Results[0].Title
			// m.Overview = r.Results[0].Overview
			// m.Video = r.Results[0].Video
			m.VoteAverage = fmt.Sprintf("%.1f", r.Results[0].VoteAverage)
			m.VoteCount = fmt.Sprint(r.Results[0].VoteCount)
		}
		// updage backdrop to english
		options["language"] = "en"
		images, imageErr := tmdbapi.tmdb.GetMovieImages(r.Results[0].ID, options)
		if imageErr == nil && len(images.Backdrops) > 0 {
			if m.BackdropPath == "" {
				m.BackdropPath = images.Backdrops[0].FilePath
			}
		}
	}

	// m.ID = int(rand.Int63())
	// m.OriginalTitle=""
	return m, nil
}

func MoviesPipelineStream(ctx context.Context, movies []*Short, tmdbkey string, limit int64) (chan *Short, chan error) {
	m, err := pipeline.Producer(ctx, movies)
	if err != nil {
		mc := make(chan *Short)
		ec := make(chan error, 1)
		close(mc)
		ec <- err
		close(ec)
		return mc, ec
	}
	mytmdb := TMDBInit(tmdbkey)
	movie_chan, errors := pipeline.Step(ctx, m, mytmdb.fetchMovieDetails, limit)
	return movie_chan, errors
}

func ChannelToMovies(ctx context.Context, cancelFunc context.CancelFunc, values <-chan *Short, errors <-chan error) ([]*Short, error) {
	movies := make([]*Short, 0)
	var firstErr error

	for values != nil || errors != nil {
		select {
		case <-ctx.Done():
			if firstErr != nil {
				return movies, firstErr
			}
			return movies, ctx.Err()
		case err, ok := <-errors:
			if !ok {
				errors = nil
				continue
			}
			if err != nil {
				cancelFunc()
				if firstErr == nil {
					firstErr = err
				}
			}
		case m, ok := <-values:
			if ok {
				if len(m.OriginalTitle) > 0 {
					m.Searchname = ""
					movies = append(movies, m)
				}
			} else {
				values = nil
			}
		}
	}

	return movies, firstErr
}
