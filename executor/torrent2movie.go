package executor

import (
	"fmt"
	"strings"

	"github.com/lieranderl/moviestracker-package/internal/movies"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type MovieConverter struct {
	// Key is "name|year" combination
	movies   map[string]*movies.Short
	torrents []*torrents.Torrent
}

func NewMovieConverter(torrents []*torrents.Torrent) *MovieConverter {
	return &MovieConverter{
		movies:   make(map[string]*movies.Short),
		torrents: torrents,
	}
}

// getMovieKey generates a unique key for name+year combination
func getMovieKey(name, year string) string {
	return fmt.Sprintf("%s|%s", name, year)
}

// Convert processes torrents and returns unique movie shorts
func (mc *MovieConverter) Convert() []*movies.Short {
	for _, torrent := range mc.torrents {
		mc.processTorrent(torrent)
	}

	return mc.getMoviesList()
}

// validYear checks if the year is valid (not a range like "2020-2021")
func isValidYear(year string) bool {
	return year != "" && !strings.Contains(year, "-")
}

// getSearchName determines the search name from torrent names
func getSearchName(torrent *torrents.Torrent) string {
	if torrent.OriginalName != "" {
		return torrent.OriginalName
	}
	return torrent.RussianName
}

// processTorrent handles a single torrent
func (mc *MovieConverter) processTorrent(torrent *torrents.Torrent) {
	// Skip if year is invalid
	if !isValidYear(torrent.Year) {
		return
	}

	searchName := getSearchName(torrent)
	movieKey := getMovieKey(searchName, torrent.Year)

	movie := mc.getOrCreateMovie(movieKey, searchName, torrent.Year)
	mc.updateMovie(movie, torrent)
}

// getOrCreateMovie retrieves existing movie or creates a new one
func (mc *MovieConverter) getOrCreateMovie(key, searchName, year string) *movies.Short {
	movie, exists := mc.movies[key]
	if !exists {
		movie = &movies.Short{
			Searchname: searchName,
			Year:       year,
			Torrents:   make([]*torrents.Torrent, 0, 1), // Pre-allocate for at least one torrent
		}
		mc.movies[key] = movie
	}
	return movie
}

// updateMovie updates movie data with torrent information
func (mc *MovieConverter) updateMovie(movie *movies.Short, torrent *torrents.Torrent) {
	movie.Torrents = append(movie.Torrents, torrent)

	// Set hash only if it's not set yet
	if movie.Hash == "" {
		movie.Hash = torrent.Hash
	}
}

// getMoviesList converts the map to a slice and finalizes movies
func (mc *MovieConverter) getMoviesList() []*movies.Short {
	result := make([]*movies.Short, 0, len(mc.movies))

	for _, movie := range mc.movies {
		movie.UpdateMoviesAttribs()
		movie.Torrents = nil // Clear torrents after processing
		result = append(result, movie)
	}

	return result
}

func (p *trackersPipeline) ConvertTorrentsToMovieShort() *trackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	converter := NewMovieConverter(p.torrents)
	p.movies = converter.Convert()

	return p
}
