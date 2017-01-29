package latest

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON_implement(t *testing.T) {
	var _ Source = &JSON{}
}

func TestJSONValidate(t *testing.T) {

	tests := []struct {
		JSON      *JSON
		expectErr bool
	}{
		{
			JSON: &JSON{
				URL: "http://good.com",
			},
			expectErr: false,
		},
		{
			JSON: &JSON{
				URL: "",
			},
			expectErr: true,
		},
	}

	for i, tt := range tests {
		j := tt.JSON
		err := j.Validate()
		if tt.expectErr == (err == nil) {
			t.Fatalf("#%d Validate() expects err == nil to eq %t", i, tt.expectErr)
		}
	}
}

// OriginalResponse implements Receiver and receives test-fixtures/original.json
type OriginalResponse struct {
	Name    string `json:"name"`
	Version string `json:"version_info"`
	Status  string `json:"status"`
}

func (r *OriginalResponse) VersionInfo() ([]string, error) {
	verStr := strings.Replace(r.Version, "v", "", 1)
	return []string{verStr}, nil
}

func (r *OriginalResponse) MetaInfo() (*Meta, error) {
	return &Meta{
		Message: r.Status,
	}, nil
}

func TestJSONFetch(t *testing.T) {

	tests := []struct {
		testServer    *httptest.Server
		response      JSONResponse
		expectCurrent string
		expectMessage string
		expectURL     string
	}{
		{
			testServer:    fakeServer("test-fixtures/default.json"),
			expectCurrent: "1.2.3",
			expectMessage: "New version include security update, you should update soon",
			expectURL:     "http://example.com/info",
		},
		{
			testServer:    fakeServer("test-fixtures/original.json"),
			expectCurrent: "1.0.0",
			expectMessage: "We are releasing now",
			response:      &OriginalResponse{},
		},
	}

	for i, tt := range tests {
		ts := tt.testServer
		defer ts.Close()

		j := &JSON{
			URL:      ts.URL,
			Response: tt.response,
		}

		fr, err := j.Fetch()
		if err != nil {
			t.Fatalf("#%d Fetch() expects error:%q to be nil", i, err.Error())
		}

		versions := fr.Versions
		current := versions[0].String()
		if current != tt.expectCurrent {
			t.Fatalf("#%d Fetch() expects %s to be %s", i, current, tt.expectCurrent)
		}

		message := fr.Meta.Message
		if message != tt.expectMessage {
			t.Fatalf("#%d Fetch() expects %q to be %q", i, message, tt.expectMessage)
		}

		url := fr.Meta.URL
		if url != tt.expectURL {
			t.Fatalf("#%d Fetch() expects %q to be %q", i, url, tt.expectURL)
		}

	}
}
