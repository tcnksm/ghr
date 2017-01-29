package latest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func fakeServer(fixture string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(fixture)
		if err != nil {
			// Should not reach here
			panic(err)
		}
		io.Copy(w, f)
	}))
}
