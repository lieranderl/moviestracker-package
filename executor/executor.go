package executor

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/lieranderl/moviestracker-package/internal/kinozal"
	"github.com/lieranderl/moviestracker-package/internal/movies"
	"github.com/lieranderl/moviestracker-package/internal/rutor"
	"github.com/lieranderl/moviestracker-package/internal/torrents"
	"github.com/lieranderl/moviestracker-package/internal/tracker"
	"github.com/lieranderl/moviestracker-package/pkg/pipeline"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type config struct {
	urls            []string
	tmdbApiKey      string
	firebaseProject string
	goption         option.ClientOption
	mongoURI        string
}

func initConfig(urls []string, tmdbkey string) *config {
	return &config{
		urls:       urls,
		tmdbApiKey: tmdbkey,
	}
}

func (c *config) WithFirebase(firebaseProject string, goption option.ClientOption) *config {
	c.firebaseProject = firebaseProject
	c.goption = goption
	return c
}

func (c *config) WithMongo(mongoURI string) *config {
	c.mongoURI = mongoURI
	return c
}

type TrackersPipeline struct {
	torrents []*torrents.Torrent
	movies   []*movies.Short
	config   config
	errors   []error
}

type firebaseEnvVars struct {
	firebaseProject string
	firebaseConfig  string
}

type mongoEnvVars struct {
	mongo_uri string
}
type EnvVars struct {
	urls       []string
	tmdbApiKey string
	firebaseEnvVars
	mongoEnvVars
}

func InitVars(urls []string, tmdbkey string) *EnvVars {
	return &EnvVars{
		urls:       urls,
		tmdbApiKey: tmdbkey,
	}
}

func (e *EnvVars) WithFirebase(firebaseProject string, firebaseConfig string) *EnvVars {
	e.mongoEnvVars = mongoEnvVars{
		mongo_uri: "",
	}
	e.firebaseEnvVars = firebaseEnvVars{
		firebaseProject: firebaseProject,
		firebaseConfig:  firebaseConfig,
	}
	return e
}

func (e *EnvVars) WithMongo(mongoUri string) *EnvVars {
	e.firebaseEnvVars = firebaseEnvVars{
		firebaseProject: "",
		firebaseConfig:  "",
	}
	e.mongoEnvVars = mongoEnvVars{
		mongo_uri: mongoUri,
	}
	return e
}

func Init(env EnvVars) *TrackersPipeline {
	tp := new(TrackersPipeline)
	tp.config = *(initConfig(env.urls, env.tmdbApiKey))
	if env.firebaseEnvVars.firebaseConfig != "" && env.firebaseEnvVars.firebaseProject != "" {
		goption := option.WithCredentialsJSON([]byte(env.firebaseConfig))
		tp.config.WithFirebase(env.firebaseEnvVars.firebaseProject, goption)

	}
	if env.mongoEnvVars.mongo_uri != "" {
		tp.config.WithMongo(env.mongoEnvVars.mongo_uri)
	}

	return tp
}

func (p *TrackersPipeline) DeleteOldMoviesFromDb() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	firestoreClient, err := firestore.NewClient(ctx, p.config.firebaseProject, p.config.goption)
	if err != nil {
		p.errors = append(p.errors, err)
	}
	moviesListRef := firestoreClient.Collection("latesttorrentsmovies").Where("LastTimeFound", "<", time.Now().Add(-time.Hour*24*30*12).Format("2006-01-02T15:04:05.000Z"))
	iter := moviesListRef.Documents(ctx)
	batch := firestoreClient.Batch()
	numDeleted := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			p.errors = append(p.errors, err)
			return p
		}

		batch.Delete(doc.Ref)
		numDeleted++
	}

	// If there are no documents to delete,
	// the process is over.
	if numDeleted == 0 {
		return p
	}

	_, err = batch.Commit(ctx)
	if err != nil {
		p.errors = append(p.errors, err)
	}
	return p
}

func (p *TrackersPipeline) ConvertTorrentsToMovieShort() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	ms := make([]*movies.Short, 0)
	hash_list := make([]string, 0)
	i := 0
	for _, movietorr := range p.torrents {
		found := false
		for _, h := range hash_list {
			if h == movietorr.Hash {
				found = true
				for _, m := range ms {
					if m.Hash == movietorr.Hash {
						m.Torrents = append(m.Torrents, movietorr)
					}
				}
				break
			}
		}
		if !found {
			hash_list = append(hash_list, movietorr.Hash)
			searchname := ""
			if movietorr.OriginalName != "" {
				searchname = movietorr.OriginalName
			} else {
				searchname = movietorr.RussianName
			}
			movie := new(movies.Short)
			// movie.Hash = movietorr.Hash
			movie.Searchname = searchname
			movie.Year = movietorr.Year
			i += 1
			movie.Torrents = append(movie.Torrents, movietorr)
			movie.UpdateMoviesAttribs()
			movie.Torrents = nil
			ms = append(ms, movie)
		}

	}
	p.movies = ms
	return p

}

func (p *TrackersPipeline) Tmdb() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	movieChan, errorChan := movies.MoviesPipelineStream(ctx, p.movies, p.config.tmdbApiKey, 20)
	p.movies = movies.ChannelToMovies(ctx, cancel, movieChan, errorChan)
	return p
}

// MONGO DB
func connectDB(mongo_uri string) *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatal(err)
	}

	//ping the database
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// getting database collections
func getCollection(client *mongo.Client, dbname, collectionName string) *mongo.Collection {
	collection := client.Database(dbname).Collection(collectionName)
	return collection
}

func (p *TrackersPipeline) SaveToDb(collection string, dbtype string) *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if dbtype == "firebase" {
		firestoreClient, err := firestore.NewClient(ctx, p.config.firebaseProject, p.config.goption)
		if err != nil {
			p.errors = append(p.errors, err)
			return p
		}
		for _, m := range p.movies {
			m.WriteMovieToFirestore(ctx, firestoreClient, collection)
		}
	}
	if dbtype == "mongo" {

		// Client instance
		DB := connectDB(p.config.mongoURI)
		moviesCol := getCollection(DB, "movies", collection)
		for _, m := range p.movies {
			m.WriteMovieToMongo(ctx, moviesCol)
		}

	}

	return p
}

func (p *TrackersPipeline) RunTrackersSearchPipilene(isMovie string) *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	var config, configKZ tracker.Config
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if isMovie == "true" {
		config = tracker.Config{Urls: []string{p.config.urls[0]}, TrackerParser: rutor.ParseMoviePage}
		configKZ = tracker.Config{Urls: []string{p.config.urls[1]}, TrackerParser: kinozal.ParseMoviePage}
	} else {
		config = tracker.Config{Urls: []string{p.config.urls[0]}, TrackerParser: rutor.ParseSeriesPage}
		configKZ = tracker.Config{Urls: []string{p.config.urls[1]}, TrackerParser: kinozal.ParseSeriesPage}
	}
	rutorTracker := tracker.Init(config)
	kzTracker := tracker.Init(configKZ)

	torrentsResults1, rutorErrors1 := rutorTracker.TorrentsPipelineStream(ctx)
	torrentsResults2, rutorErrors2 := kzTracker.TorrentsPipelineStream(ctx)

	allTorrents := pipeline.Merge(ctx, torrentsResults1, torrentsResults2)
	allErrors := pipeline.Merge(ctx, rutorErrors1, rutorErrors2)

	ts, err := torrents.MergeTorrentChannlesToSlice(ctx, cancel, allTorrents, allErrors)
	log.Println("TS length before: ", len(ts))
	ts = torrents.RemoveDuplicatesInPlace(ts)
	log.Println("TS length AFTER: ", len(ts))

	if err != nil {
		p.errors = append(p.errors, err)
	} else {
		p.torrents = ts
	}
	return p
}

func (p *TrackersPipeline) RunRutorPipiline() *TrackersPipeline {
	if len(p.errors) > 0 {
		return p
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := tracker.Config{Urls: p.config.urls, TrackerParser: rutor.ParseMoviePage}
	rutorTracker := tracker.Init(config)
	torrentsResults, rutorErrors := rutorTracker.TorrentsPipelineStream(ctx)

	ts, err := torrents.MergeTorrentChannlesToSlice(ctx, cancel, torrentsResults, rutorErrors)
	if err != nil {
		p.errors = append(p.errors, err)
	} else {
		p.torrents = ts
	}
	return p
}

func (p *TrackersPipeline) HandleErrors() error {
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
