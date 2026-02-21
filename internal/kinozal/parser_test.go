package kinozal

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration test skipped: set RUN_INTEGRATION_TESTS=1 to enable")
	}

	_ = godotenv.Load("../../.env")
	if os.Getenv("KZ_LOGIN") == "" || os.Getenv("KZ_PASSWORD") == "" {
		t.Skip("integration test skipped: KZ_LOGIN/KZ_PASSWORD are not configured")
	}

	name := "Вышка"
	year := "2022"
	test_link := "http://kinozal.tv/browse.php?s=" + name + "%281080p%7C2160p%29&g=3&c=0&v=0&d=" + year + "&w=0&t=0&f=0"
	tors, err := ParseMoviePage(test_link)

	if err != nil {
		t.Fatalf("ParseMoviePage failed: %v", err)
	}

	for _, tor := range tors {
		if tor.DetailsUrl == "1933025" {
			assert.Equal(t, "magnet:?xt=urn:btih:A8BABAB89BE27F66141477DF58096ECC9721CF59", tor.Magnet)
			return
		}
	}

	t.Fatalf("expected torrent with details id 1933025 not found")
}
