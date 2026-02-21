package torrents

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDuplicatesInPlace(t *testing.T) {
	input := []*Torrent{
		{Name: "A1", MagnetHash: "bbb"},
		{Name: "A2", MagnetHash: "AAA"},
		{Name: "A3", MagnetHash: "aaa"},
		{Name: "A4", MagnetHash: "ccc"},
	}

	got := RemoveDuplicatesInPlace(input)

	require.Len(t, got, 3)
	require.Equal(t, "AAA", got[0].MagnetHash)
	require.Equal(t, "bbb", got[1].MagnetHash)
	require.Equal(t, "ccc", got[2].MagnetHash)
}

func TestMergeTorrentChannelsToSliceSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	values := make(chan []*Torrent, 1)
	errs := make(chan error, 1)

	values <- []*Torrent{{Name: "movie-1"}, {Name: "movie-2"}}
	close(values)
	close(errs)

	got, err := MergeTorrentChannelsToSlice(ctx, cancel, values, errs)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestMergeTorrentChannelsToSliceError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	values := make(chan []*Torrent)
	errs := make(chan error, 1)

	wantErr := errors.New("boom")
	errs <- wantErr
	close(errs)
	close(values)

	got, err := MergeTorrentChannelsToSlice(ctx, cancel, values, errs)
	require.Error(t, err)
	require.Equal(t, wantErr, err)
	require.Empty(t, got)
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestMergeTorrentChannlesToSliceCompatibilityWrapper(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	values := make(chan []*Torrent, 1)
	errs := make(chan error)

	values <- []*Torrent{{Name: "movie-1"}}
	close(values)
	close(errs)

	got, err := MergeTorrentChannlesToSlice(ctx, cancel, values, errs)
	require.NoError(t, err)
	require.Len(t, got, 1)
}
