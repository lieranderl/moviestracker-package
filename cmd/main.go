// Demo
package main

import (
	"context"
	"log"
	"strings"

	"os"

	"time"

	"github.com/joho/godotenv"
	"github.com/lieranderl/moviestracker-package/executor"
)

func main() {

	start := time.Now()
	godotenv.Load()

	// urls := []string{os.Getenv("RUTOR_SEARCH_URL")}
	// 	fmt.Sprintf(os.Getenv("KZ_SEARCH_URL"), "Bad boys", "2024")}

	rutorUrls := strings.Split(os.Getenv("RUTOR_DV_URLS"), ",")
	kinizalUrls := strings.Split(os.Getenv("KINOZAL_DV_URLS"), ",")
	urls := append(rutorUrls, kinizalUrls...)

	// Initialize the pipeline
	pipeline := executor.InitPipeline(context.Background(), urls, os.Getenv("TMDBAPIKEY"), executor.OptionWithMongo(os.Getenv("MONGO_URI")))

	err := pipeline.
		RunTrackersSearchPipeline(true).
		ConvertTorrentsToMovieShort().
		Tmdb().
		// SaveToDb().
		HandleErrors()

	log.Println(err)

	log.Println(pipeline)

	// pring all collected movies
	for _, m := range pipeline.GetMovies() {
		log.Println("Movie:", m.ID, m.Title, m.OriginalTitle, m.Year, m.VoteAverage)
	}

	elapsed := time.Since(start)
	log.Printf("ALL took %s", elapsed)

}
