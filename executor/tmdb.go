package executor

import (
	"context"

	"github.com/lieranderl/moviestracker-package/internal/movies"
)

func (p *trackersPipeline) Tmdb() *trackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	ctx, cancel := context.WithCancel(p.config.ctx)
	defer cancel()
	movieChan, errorChan := movies.MoviesPipelineStream(ctx, p.movies, p.config.tmdbApiKey, 20)
	p.movies = movies.ChannelToMovies(ctx, cancel, movieChan, errorChan)

	return p
}
