package executor

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/lieranderl/moviestracker-package/internal/movies"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/lieranderl/moviestracker-package/internal/trackers"
	"github.com/lieranderl/moviestracker-package/pkg/pipeline"

	"google.golang.org/api/option"
)

// Config holds all configuration for the pipeline.
type config struct {
	ctx             context.Context
	urls            []string
	tmdbApiKey      string
	firebaseProject string
	firebaseConfig  string
	mongoURI        string
	goption         option.ClientOption
}

// NewConfig initializes a basic configuration.
func newConfig(ctx context.Context, urls []string, tmdbApiKey string) *config {
	return &config{
		ctx:        ctx,
		urls:       urls,
		tmdbApiKey: tmdbApiKey,
	}
}

// WithFirebase adds Firebase configuration if provided.
func (c *config) withFirebase(project, config string) *config {
	if project != "" && config != "" {
		c.firebaseProject = project
		c.firebaseConfig = config
		c.goption = option.WithCredentialsJSON([]byte(config))
	}
	return c
}

// WithMongo adds MongoDB configuration if provided.
func (c *config) withMongo(uri string) *config {
	if uri != "" {
		c.mongoURI = uri
	}
	return c
}

// TrackersPipeline orchestrates operations for torrents and movies.
type trackersPipeline struct {
	torrents []*torrents.Torrent
	movies   []*movies.Short
	config   *config
	errors   []error
}

// NewTrackersPipeline initializes a pipeline with the given config.
func newTrackersPipeline(config *config) *trackersPipeline {
	return &trackersPipeline{
		config: config,
	}
}

func InitPipeline(ctx context.Context, urls []string, tmdbApiKey string, options ...func(*config)) *trackersPipeline {
	// Initialize the base config
	config := newConfig(ctx, urls, tmdbApiKey)

	// Apply optional configurations
	for _, option := range options {
		option(config)
	}

	return newTrackersPipeline(config)
}

// OptionWithFirebase configures Firebase settings.
func OptionWithFirebase(project, firebase_config string) func(*config) {
	return func(c *config) {
		c.withFirebase(project, firebase_config)
	}
}

// OptionWithMongo configures MongoDB settings.
func OptionWithMongo(uri string) func(*config) {
	return func(c *config) {
		c.withMongo(uri)
	}
}

type PipelineConfig struct {
	Parsers []trackers.ParserConfig
	IsMovie bool
}

// GetTorrents returns the list of torrents.
func (p *trackersPipeline) GetTorrents() []*torrents.Torrent {
	return p.torrents
}

// GetMovies returns the list of movies.
func (p *trackersPipeline) GetMovies() []*movies.Short {
	return p.movies
}

func (p *trackersPipeline) RunTrackersSearchPipeline(isMovie bool) *trackersPipeline {
	configs := []trackers.ParserConfig{
		trackers.NewRutorConfig("http://rutor.is"),
		trackers.NewKinozalConfig("https://kinozal.tv"),
	}

	processor, err := trackers.NewTrackerProcessor(isMovie, configs...)
	if err != nil {
		p.errors = append(p.errors, err)
	}

	log.Println("URLS:", p.config.urls)

	urlStream, err := pipeline.Producer(p.config.ctx, p.config.urls)
	if err != nil {
		p.errors = append(p.errors, err)
	}

	torrentsChan, errorsChan := pipeline.Step(
		p.config.ctx,
		urlStream,
		processor.ProcessURL,
		5,
	)


	var allTorrents []*torrents.Torrent
	var errs []error

	done := make(chan struct{})
	go func() {
		defer close(done)
		for err := range errorsChan {
			errs = append(errs, err)
		}
	}()

	for torrents := range torrentsChan {
		allTorrents = append(allTorrents, torrents...)
	}

	<-done

	p.torrents = torrents.RemoveDuplicatesInPlace(allTorrents)
	return p
}

func (p *trackersPipeline) HandleErrors() error {
	var err error
	if len(p.errors) > 0 {
		errorStrSlice := make([]string, 0)
		for _, err := range p.errors {
			errorStrSlice = append(errorStrSlice, err.Error())
		}
		err := errors.New(strings.Join(errorStrSlice, ",\n"))
		log.Println(err)
	}
	return err
}
