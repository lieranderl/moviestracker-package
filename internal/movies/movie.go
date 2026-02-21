package movies

import (
	"time"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type Short struct {
	BackdropPath  string  `json:"backdrop_path" bson:"backdrop_path"`
	ID            string  `json:"id" bson:"id"`
	OriginalTitle string  `json:"original_title" bson:"original_title"`
	GenreIDs      []int32 `json:"genre_ids" bson:"genre_ids"`
	PosterPath    string  `json:"poster_path" bson:"poster_path"`
	ReleaseDate   string  `json:"release_date" bson:"release_date"`
	Title         string  `json:"title"`
	VoteAverage   string  `json:"vote_average" bson:"vote_average"`
	VoteCount     string  `json:"vote_count" bson:"vote_count"`
	Year          string
	Torrents      []*torrents.Torrent `json:"torrents" bson:"torrents,omitempty"`
	Hash          string              `json:"hash" bson:"hash,omitempty"`
	Searchname    string              `json:"searchname" bson:"searchname,omitempty"`
	LastTimeFound time.Time           `json:"lasttimefound" bson:"lasttimefound"`
}

func (m *Short) UpdateMoviesAttribs() {
	for _, t := range m.Torrents {
		m.setLastTimeFound(t)
	}
}

func (m *Short) setLastTimeFound(t *torrents.Torrent) {
	const layout = "2006-01-02T15:04:05.000Z"

	if t.Date == "" {
		t.Date = time.Now().UTC().Format(layout)
	}

	timeformat, err := time.Parse(layout, t.Date)
	if err != nil {
		timeformat = time.Now().UTC()
	}

	if timeformat.After(m.LastTimeFound) {
		m.LastTimeFound = timeformat
	}
}
