package movies

import (
	"time"

	"github.com/lieranderl/moviestracker-package/internal/torrents"
)

type Short struct {
	// Adult         bool    `json:"adult,omitempty" firestore:"adult,omitempty"`
	BackdropPath  string  `json:"backdrop_path" firestore:"backdrop_path" bson:"backdrop_path"`
	ID            string  `json:"id" firestore:"id" bson:"id"`
	OriginalTitle string  `json:"original_title" firestore:"original_title" bson:"original_title"`
	GenreIDs      []int32 `json:"genre_ids" firestore:"genre_ids" bson:"genre_ids"`
	// Popularity    float32 `json:"popularity" firestore:"popularity"`
	PosterPath  string `json:"poster_path" firestore:"poster_path" bson:"poster_path"`
	ReleaseDate string `json:"release_date" firestore:"release_date" bson:"release_date"`
	Title       string `json:"title" firestore:"title"`
	// Overview      string  `json:"overview" firestore:"overview"`
	// Video         bool    `json:"video" firestore:"video"`
	VoteAverage string `json:"vote_average" firestore:"vote_average" bson:"vote_average"`
	VoteCount   string `json:"vote_count" firestore:"vote_count" bson:"vote_count"`
	Year        string
	Torrents    []*torrents.Torrent `json:"torrents" firestore:"torrents" bson:"torrents,omitempty"`
	Hash        string              `json:"hash" firestore:"hash" bson:"hash,omitempty"`
	Searchname  string              `json:"searchname" firestore:"searchname" bson:"searchname,omitempty"`
	// K4            bool
	// FHD           bool
	// HDR           bool
	// HDR10         bool
	// HDR10plus     bool
	// DV            bool
	LastTimeFound time.Time `json:"lasttimefound" firestore:"lasttimefound" bson:"lasttimefound"`
}

func (m *Short) UpdateMoviesAttribs() {
	for _, t := range m.Torrents {
		// m.setQualityVector(t)
		m.setLastimeFound(t)
	}
}

// func (m *Short) setQualityVector(t *torrents.Torrent) {
// 	if t.K4 {
// 		m.K4 = true
// 	}
// 	if t.FHD {
// 		m.FHD = true
// 	}
// 	if t.HDR {
// 		m.HDR = true
// 	}
// 	if t.HDR10 {
// 		m.HDR10 = true
// 	}
// 	if t.HDR10plus {
// 		m.HDR10plus = true
// 	}
// 	if t.DV {
// 		m.DV = true
// 	}
// }

func (m *Short) setLastimeFound(t *torrents.Torrent) {
	if t.Date == "" {
		t.Date = time.Now().String()
	}

	layout := "2006-01-02"
	timeformat, _ := time.Parse(layout, t.Date)

	if timeformat.After(m.LastTimeFound) {
		m.LastTimeFound = timeformat
	}
}
