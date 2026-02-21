package executor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/lieranderl/moviestracker-package/internal/kinozal"
	"github.com/lieranderl/moviestracker-package/internal/movies"
	"github.com/lieranderl/moviestracker-package/internal/rutor"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/lieranderl/moviestracker-package/internal/tracker"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DBTypeMongo = "mongo"
	mongoDBName = "movies"
)

type config struct {
	urls       []string
	tmdbAPIKey string
	mongoURI   string
}

func initConfig(urls []string, tmdbKey string) *config {
	return &config{
		urls:       append([]string(nil), urls...),
		tmdbAPIKey: tmdbKey,
	}
}

func (c *config) WithMongo(mongoURI string) *config {
	c.mongoURI = strings.TrimSpace(mongoURI)
	return c
}

type TrackersPipeline struct {
	torrents []*torrents.Torrent
	movies   []*movies.Short
	config   config
	errors   []error
}

func (p *TrackersPipeline) GetTorrents() []*torrents.Torrent {
	return p.torrents
}

type EnvVars struct {
	urls       []string
	tmdbAPIKey string
	mongoURI   string
}

func InitVars(urls []string, tmdbKey string) *EnvVars {
	return &EnvVars{
		urls:       append([]string(nil), urls...),
		tmdbAPIKey: tmdbKey,
	}
}

func (e *EnvVars) WithMongo(mongoURI string) *EnvVars {
	e.mongoURI = strings.TrimSpace(mongoURI)
	return e
}

func Init(env EnvVars) *TrackersPipeline {
	tp := new(TrackersPipeline)
	tp.config = *(initConfig(env.urls, env.tmdbAPIKey))

	if env.mongoURI != "" {
		tp.config.WithMongo(env.mongoURI)
	}

	return tp
}

func (p *TrackersPipeline) addError(err error) {
	if err != nil {
		p.errors = append(p.errors, err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (p *TrackersPipeline) ConvertTorrentsToMovieShort() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	grouped := make(map[string]*movies.Short)
	order := make([]string, 0, len(p.torrents))

	for _, movieTorrent := range p.torrents {
		if movieTorrent == nil {
			continue
		}

		hash := firstNonEmpty(movieTorrent.Hash, movieTorrent.MagnetHash, movieTorrent.Name)
		if hash == "" {
			continue
		}

		movie, found := grouped[hash]
		if !found {
			movie = &movies.Short{
				Hash:       hash,
				Searchname: firstNonEmpty(movieTorrent.OriginalName, movieTorrent.RussianName, movieTorrent.Name),
				Year:       movieTorrent.Year,
				Torrents:   make([]*torrents.Torrent, 0, 1),
			}
			grouped[hash] = movie
			order = append(order, hash)
		} else if movie.Year == "" && movieTorrent.Year != "" {
			movie.Year = movieTorrent.Year
		}

		movie.Torrents = append(movie.Torrents, movieTorrent)
	}

	shortMovies := make([]*movies.Short, 0, len(grouped))
	for _, hash := range order {
		movie := grouped[hash]
		movie.UpdateMoviesAttribs()
		movie.Torrents = nil
		shortMovies = append(shortMovies, movie)
	}

	p.movies = shortMovies
	return p
}

func (p *TrackersPipeline) Tmdb() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	slog.Info("tmdb enrichment started", "movies", len(p.movies))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	movieChan, errorChan := movies.MoviesPipelineStream(ctx, p.movies, p.config.tmdbAPIKey, 20)
	enrichedMovies, err := movies.ChannelToMovies(ctx, cancel, movieChan, errorChan)
	if err != nil {
		p.addError(err)
		return p
	}

	p.movies = enrichedMovies
	slog.Info("tmdb enrichment completed", "movies", len(p.movies))
	return p
}

func connectMongo(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(50)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

func (p *TrackersPipeline) SaveToMongo(collection string) *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	if p.config.mongoURI == "" {
		p.addError(errors.New("mongo uri is required to save movies"))
		return p
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	slog.Info("mongodb save started", "collection", collection, "movies", len(p.movies))

	client, err := connectMongo(ctx, p.config.mongoURI)
	if err != nil {
		p.addError(err)
		return p
	}
	defer func() {
		p.addError(client.Disconnect(context.Background()))
	}()

	moviesCollection := client.Database(mongoDBName).Collection(collection)
	for _, movie := range p.movies {
		p.addError(movie.WriteMovieToMongo(ctx, moviesCollection))
	}

	slog.Info("mongodb save finished", "collection", collection, "movies", len(p.movies))

	return p
}

func (p *TrackersPipeline) SaveToDb(collection string, dbType string) *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	switch strings.TrimSpace(strings.ToLower(dbType)) {
	case "", DBTypeMongo:
		return p.SaveToMongo(collection)
	default:
		p.addError(fmt.Errorf("unsupported db type %q (only %q is supported)", dbType, DBTypeMongo))
		return p
	}
}

func (p *TrackersPipeline) RunTrackersSearchPipeline(isMovie bool) *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	if len(p.config.urls) < 2 {
		p.addError(errors.New("at least two tracker urls are required: rutor and kinozal"))
		return p
	}

	var rutorParser func(string) ([]*torrents.Torrent, error)
	var kinozalParser func(string) ([]*torrents.Torrent, error)

	if isMovie {
		rutorParser = rutor.ParseMoviePage
		kinozalParser = kinozal.ParseMoviePage
	} else {
		rutorParser = rutor.ParseSeriesPage
		kinozalParser = kinozal.ParseSeriesPage
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rutorTracker := tracker.Init(tracker.Config{
		Urls:          []string{p.config.urls[0]},
		TrackerParser: rutorParser,
	})
	kinozalTracker := tracker.Init(tracker.Config{
		Urls:          []string{p.config.urls[1]},
		TrackerParser: kinozalParser,
	})

	rutorTorrents, rutorErrors := rutorTracker.TorrentsPipelineStream(ctx)
	kinozalTorrents, kinozalErrors := kinozalTracker.TorrentsPipelineStream(ctx)

	var (
		rutorResult   []*torrents.Torrent
		kinozalResult []*torrents.Torrent
		rutorErr      error
		kinozalErr    error
		wg            sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		rutorResult, rutorErr = torrents.MergeTorrentChannelsToSlice(ctx, cancel, rutorTorrents, rutorErrors)
	}()
	go func() {
		defer wg.Done()
		kinozalResult, kinozalErr = torrents.MergeTorrentChannelsToSlice(ctx, cancel, kinozalTorrents, kinozalErrors)
	}()
	wg.Wait()

	if rutorErr != nil && !errors.Is(rutorErr, context.Canceled) {
		p.addError(rutorErr)
	}
	if kinozalErr != nil && !errors.Is(kinozalErr, context.Canceled) {
		p.addError(kinozalErr)
	}
	if len(p.errors) > 0 {
		return p
	}

	slog.Info(
		"tracker results",
		"rutor_torrents", len(rutorResult),
		"kinozal_torrents", len(kinozalResult),
	)

	ts := append(rutorResult, kinozalResult...)
	before := len(ts)
	ts = torrents.RemoveDuplicatesInPlace(ts)
	slog.Info("tracker search completed", "torrents_before_dedupe", before, "torrents_after_dedupe", len(ts))

	p.torrents = ts
	return p
}

func (p *TrackersPipeline) RunTrackersSearchPipilene(isMovie string) *TrackersPipeline {
	return p.RunTrackersSearchPipeline(strings.EqualFold(strings.TrimSpace(isMovie), "true"))
}

func (p *TrackersPipeline) RunRutorPipeline() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	if len(p.config.urls) == 0 {
		p.addError(errors.New("at least one rutor url is required"))
		return p
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rutorTracker := tracker.Init(tracker.Config{
		Urls:          p.config.urls,
		TrackerParser: rutor.ParseMoviePage,
	})

	torrentsResults, rutorErrors := rutorTracker.TorrentsPipelineStream(ctx)
	ts, err := torrents.MergeTorrentChannelsToSlice(ctx, cancel, torrentsResults, rutorErrors)
	if err != nil {
		p.addError(err)
		return p
	}

	p.torrents = ts
	return p
}

func (p *TrackersPipeline) RunRutorPipiline() *TrackersPipeline {
	return p.RunRutorPipeline()
}

func (p *TrackersPipeline) HandleErrors() error {
	if len(p.errors) == 0 {
		return nil
	}

	errorStrSlice := make([]string, 0, len(p.errors))
	for _, err := range p.errors {
		if err != nil {
			errorStrSlice = append(errorStrSlice, err.Error())
		}
	}

	if len(errorStrSlice) == 0 {
		return nil
	}

	return errors.New(strings.Join(errorStrSlice, ",\n"))
}
