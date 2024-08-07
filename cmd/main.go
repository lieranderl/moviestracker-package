// Demo
package main

import (
	"log"
	"os"
	// "strings"
	"time"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/lieranderl/moviestracker-package/executor"
)



func main() {
	
	// if len(os.Args) < 2 {
	// 	log.Println("No arguments provided")
	// 	return
	// }
	godotenv.Load()
	// dbname := os.Args[1]
	// dbtype := os.Args[2]
	// urls := ""
	// switch dbname {
	// case "latesttorrentsmovies":
	// 	urls = os.Getenv("RUTOR_URLS")
	// case "hdr10movies":
	// 	urls = os.Getenv("RUTOR_HDR10_URLS")
	// case "dvmovies":
	// 	urls = os.Getenv("RUTOR_DV_URLS")
	// default:
	// 	log.Println("Wrong argument provided")
	// 	return
	// }

	// log.Printf("Start %s !", dbname)
	start := time.Now()

	//	latesttorrentsmovies	strings.Split(os.Getenv("RUTOR_URLS"), ","),
	//	hdr10movies	strings.Split(os.Getenv("RUTOR_HDR10_URLS"), ","),
	//	dvmovies	strings.Split(os.Getenv("RUTOR_DV_URLS"), ","),
	
	// env_vars := executor.InitVars(strings.Split(urls, ","), os.Getenv("TMDBAPIKEY"))
	// switch dbtype {
	// case "mongo":
	// 	env_vars.WithMongo(os.Getenv("MONGO_URI"))
	// case "firebase":
	// 	env_vars.WithFirebase(os.Getenv("FIREBASE_PROJECT"), os.Getenv("FIREBASECONFIG"))
	// }
	// pipeline := executor.Init(
		// *env_vars,
	// )
	// err := pipeline.
	// 	RunRutorPipiline(). 
	// 	ConvertTorrentsToMovieShort().
	// 	Tmdb().
	// 	SaveToDb(dbname, dbtype).
	// 	// DeleteOldMoviesFromDb().
		// HandleErrors()

	// err := pipeline.RunTrackersSearchPipilene("true")
	// if err != nil {
	// 	return
	// }

	env_vars := executor.InitVars([]string{
		fmt.Sprintf(os.Getenv("RUTOR_SEARCH_URL"), "Bad boys", "2024"),
		fmt.Sprintf(os.Getenv("KZ_SEARCH_URL"), "Bad boys", "2024"),
	}, os.Getenv("TMDBAPIKEY"))
	pipeline := executor.Init(*env_vars)
	err := pipeline.RunTrackersSearchPipilene("true").HandleErrors()

	log.Println(err)

	log.Println(pipeline)

	elapsed := time.Since(start)
	log.Printf("ALL took %s", elapsed)

}
