package rutor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsRussianOnly(t *testing.T) {
	require.True(t, isRussianOnly("Плохие парни"))
	require.False(t, isRussianOnly("Плохие парни / Bad Boys"))
}

func TestParseDate(t *testing.T) {
	got := parseDate("06 Янв 24")
	require.Equal(t, "2024-01-06T00:00:00.000Z", got)
}

func TestParseDateFallback(t *testing.T) {
	got := parseDate("invalid")

	_, err := time.Parse("2006-01-02T15:04:05.000Z", got)
	require.NoError(t, err)
}
