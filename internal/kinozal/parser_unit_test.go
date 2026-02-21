package kinozal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDetailsID(t *testing.T) {
	require.Equal(t, "1933366", parseDetailsID("/details.php?id=1933366"))
	require.Equal(t, "", parseDetailsID("/details.php"))
}

func TestParseTitleMetadata(t *testing.T) {
	torrent := &kzTorrent{}
	torrent.Name = "Плохие парни до конца / Bad Boys: Ride or Die (2024) WEB-DL 2160p"

	torrent.parseTitleMetadata()

	require.Equal(t, "2024", torrent.Year)
	require.Equal(t, "Плохие парни до конца", torrent.RussianName)
	require.Equal(t, "Bad Boys: Ride or Die", torrent.OriginalName)
	require.NotEmpty(t, torrent.Hash)
}
