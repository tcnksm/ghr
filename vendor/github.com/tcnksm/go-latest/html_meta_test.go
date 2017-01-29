package latest

import (
	"net/http/httptest"
	"testing"
)

func TestHTMLMeta_implement(t *testing.T) {
	var _ Source = &HTMLMeta{}
}

func TestHTMLMetaFetch(t *testing.T) {
	tests := []struct {
		name          string
		testServer    *httptest.Server
		expectCurrent string
		expectMessage string
	}{
		{
			name:          "reduce-worker",
			testServer:    fakeServer("test-fixtures/meta.html"),
			expectCurrent: "1.2.1",
			expectMessage: "New version include security update",
		},
	}

	for i, tt := range tests {
		ts := tt.testServer
		defer ts.Close()

		h := &HTMLMeta{
			URL:  ts.URL,
			Name: tt.name,
		}

		fr, err := h.Fetch()
		if err != nil {
			t.Fatalf("#%d Fetch() expects error:%q to be nil", i, err.Error())
		}

		versions := fr.Versions
		if len(versions) == 0 {
			t.Fatalf("#%d Fetch() expects number of versions found from HTML not to be 0", i)
		}

		current := versions[0].String()
		if current != tt.expectCurrent {
			t.Fatalf("#%d Fetch() expects %s to be %s", i, current, tt.expectCurrent)
		}

		message := fr.Meta.Message
		if message != tt.expectMessage {
			t.Fatalf("#%d Fetch() expects %q to be %q", i, message, tt.expectMessage)
		}
	}

}
