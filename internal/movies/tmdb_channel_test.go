package movies

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelToMoviesCollectsValidMoviesAndReturnsFirstError(t *testing.T) {
	ctx := context.Background()
	cancel := context.CancelFunc(func() {})

	values := make(chan *Short, 2)
	errs := make(chan error, 2)

	values <- &Short{OriginalTitle: "Movie A"}
	values <- &Short{OriginalTitle: ""}
	close(values)

	wantErr := errors.New("tmdb failed")
	errs <- wantErr
	close(errs)

	got, err := ChannelToMovies(ctx, cancel, values, errs)

	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
	require.Len(t, got, 1)
	require.Equal(t, "Movie A", got[0].OriginalTitle)
}
