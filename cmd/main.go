package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lieranderl/moviestracker-package/executor"
	"github.com/lieranderl/moviestracker-package/pkg/logging"
)

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildTrackerURL(template, query, year string) (string, error) {
	raw := fmt.Sprintf(template, query, year)

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse tracker url %q: %w", raw, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New("tracker url must include scheme and host")
	}

	// Kinozal currently responds reliably over HTTPS.
	if strings.EqualFold(u.Host, "kinozal.tv") && strings.EqualFold(u.Scheme, "http") {
		u.Scheme = "https"
	}

	if u.RawQuery != "" {
		q := u.Query()
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

func main() {
	logger := logging.Init()
	_ = godotenv.Load()

	var (
		query       = flag.String("query", envOrDefault("MOVIE_QUERY", "Bad Boys"), "Movie/series search query")
		year        = flag.String("year", envOrDefault("MOVIE_YEAR", "2024"), "Release year")
		isMovie     = flag.Bool("movie", true, "Set false to search for series")
		saveToMongo = flag.Bool("save", strings.EqualFold(envOrDefault("SAVE_TO_MONGO", "false"), "true"), "Persist enriched movies to MongoDB")
		collection  = flag.String("collection", envOrDefault("MONGO_COLLECTION", "movies"), "MongoDB collection name")
	)
	flag.Parse()

	rutorSearchURL := os.Getenv("RUTOR_SEARCH_URL")
	kinozalSearchURL := os.Getenv("KZ_SEARCH_URL")
	tmdbAPIKey := os.Getenv("TMDBAPIKEY")
	mongoURI := os.Getenv("MONGO_URI")

	if rutorSearchURL == "" || kinozalSearchURL == "" {
		logger.Error("missing required tracker urls", "required", []string{"RUTOR_SEARCH_URL", "KZ_SEARCH_URL"})
		os.Exit(1)
	}
	if tmdbAPIKey == "" {
		logger.Error("missing required tmdb api key", "required", "TMDBAPIKEY")
		os.Exit(1)
	}

	rutorURL, err := buildTrackerURL(rutorSearchURL, *query, *year)
	if err != nil {
		logger.Error("failed to build rutor url", "error", err)
		os.Exit(1)
	}
	kinozalURL, err := buildTrackerURL(kinozalSearchURL, *query, *year)
	if err != nil {
		logger.Error("failed to build kinozal url", "error", err)
		os.Exit(1)
	}
	urls := []string{rutorURL, kinozalURL}

	start := time.Now()
	logger.Info("starting tracker pipeline", "query", *query, "year", *year, "is_movie", *isMovie)
	logger.Debug("tracker urls", "rutor_url", rutorURL, "kinozal_url", kinozalURL)

	envVars := executor.InitVars(urls, tmdbAPIKey)
	if mongoURI != "" {
		envVars.WithMongo(mongoURI)
	}
	pipeline := executor.Init(*envVars)

	pipeline = pipeline.
		RunTrackersSearchPipeline(*isMovie).
		ConvertTorrentsToMovieShort().
		Tmdb()

	if *saveToMongo {
		if mongoURI == "" {
			logger.Error("mongo save enabled but MONGO_URI is missing")
			os.Exit(1)
		}
		pipeline = pipeline.SaveToMongo(*collection)
	}

	err = pipeline.HandleErrors()
	if err != nil {
		logger.Error("pipeline failed", "error", err)
		os.Exit(1)
	}

	logger.Info("pipeline completed", "torrents", len(pipeline.GetTorrents()), "elapsed", time.Since(start).String())
}
