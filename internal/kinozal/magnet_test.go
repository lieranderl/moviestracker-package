package kinozal

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestMagnet(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("integration test skipped: set RUN_INTEGRATION_TESTS=1 to enable")
	}

	_ = godotenv.Load("../../.env")
	if os.Getenv("KZ_LOGIN") == "" || os.Getenv("KZ_PASSWORD") == "" {
		t.Skip("integration test skipped: KZ_LOGIN/KZ_PASSWORD are not configured")
	}

	cred := new(Cred)
	cred.Login = os.Getenv("KZ_LOGIN")
	cred.Password = os.Getenv("KZ_PASSWORD")

	var jar *cookiejar.Jar
	tbTransport := &http.Transport{}
	jar, _ = cookiejar.New(&cookiejar.Options{})
	httpcl := &http.Client{Jar: jar, Transport: tbTransport, Timeout: defaultRequestTimeout}

	l, httpcl := Login(httpcl, cred)
	if l {
		ch := make(chan map[string]string)
		go GetMagnet(httpcl, "1933366", ch)
		yy := <-ch
		close(ch)
		assert.Equal(t, "magnet:?xt=urn:btih:A14E679DD461CF6A3C70FAACAB2EEC95C66AA817", yy["1933366"])
	} else {
		t.Fatal("failed to login to kinozal")
	}

}
