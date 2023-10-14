package cloudconvert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCreateConversionTasks(t *testing.T) {
	t.Parallel()
	cloudConvertToken := "expected-token"

	urls := []string{
		"https://storage/1.pdf",
		"https://storage/2.pdf",
		"https://storage/3.pdf",
		"https://storage/4 5.pdf",     // filename is "4 5.pdf" (contains space)
		"https://storage/k=6&v=7.pdf", // filename is "k=6&v=7.pdf"
		"https://storage/.特殊文字.pdf",   // filename is ".特殊文字.pdf"
	}
	urlsMapUUID := make(map[string]string)
	for idx, u := range urls {
		p, _ := url.Parse(u)
		urlsMapUUID[p.EscapedPath()] = strconv.Itoa(idx + 1)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.Contains(authz, cloudConvertToken) {
			t.Errorf("missing token")
		}

		var body tasks
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}

		var gotURL string
		for k, v := range body.Tasks {
			if strings.Contains(k, "import-") {
				gotURL = v.URL
				break
			}
		}

		var found bool
		var parsedURL *url.URL
		for _, u := range urls {
			parsedURL, _ = url.Parse(u)
			if gotURL == parsedURL.String() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected url: %q", gotURL)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createJobResponse{
			Data: struct {
				ID string `json:"id"`
			}{
				ID: urlsMapUUID[parsedURL.EscapedPath()],
			},
		})
	}))
	defer ts.Close()

	svc := Service{
		Host:   ts.URL,
		Token:  cloudConvertToken,
		Client: http.DefaultClient,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := svc.CreateConversionTasks(ctx, urls)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(tasks) != len(urls) {
		t.Errorf("total tasks must match total urls")
	}

	// make sure tasks have the same order with urls
	for i, task := range tasks {
		p, _ := url.Parse(urls[i])
		expectedUUID := urlsMapUUID[p.EscapedPath()]
		if expectedUUID != task {
			t.Errorf("task UUID does not match, got: %q, want: %q", task, expectedUUID)
		}
	}
}

func TestUploadPrefixURL(t *testing.T) {
	t.Parallel()

	endpoint := "endpoint"
	bucket := "bucket"
	s := &Service{StorageBucket: bucket, StorageEndpoint: endpoint}

	want := endpoint + "/" + bucket
	if got := s.UploadPrefixURL(); got != want {
		t.Errorf("s.UploadPrefixURL() = %q, want %q", got, want)
	}
}
