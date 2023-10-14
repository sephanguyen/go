package whiteboard_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
)

func TestCreateRoom(t *testing.T) {
	t.Parallel()
	expectedRoomUUID := "uuid"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(struct {
			UUID string `json:"uuid"`
		}{
			UUID: expectedRoomUUID,
		})
	}))
	defer ts.Close()

	svc := whiteboard.New(&configs.WhiteboardConfig{
		Endpoint: ts.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	room, err := svc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
		Name: "room name",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if room.UUID != expectedRoomUUID {
		t.Errorf("unexpected room uuid, got: %q, want: %q", room.UUID, expectedRoomUUID)
	}
}

func TestFetchRoomToken(t *testing.T) {
	t.Parallel()
	expectedToken := "token"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`"` + expectedToken + `"`))
	}))
	defer ts.Close()

	svc := whiteboard.New(&configs.WhiteboardConfig{
		Endpoint: ts.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := svc.FetchRoomToken(ctx, "room")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token != expectedToken {
		t.Errorf("unexpected room token, got: %q, want: %q", token, expectedToken)
	}
}

func TestCreateConversionTasks(t *testing.T) {
	t.Parallel()
	urls := []string{
		"https://1",
		"https://2",
		"https://3",
		"https://4",
		"https://5",
	}
	urlsMapUUID := make(map[string]string)
	for _, url := range urls {
		urlsMapUUID[url] = strings.Replace(url, "https://", "", -1)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Resource string
			Type     string
		}
		json.NewDecoder(r.Body).Decode(&request)

		if request.Type != "static" {
			t.Errorf("request type must be %q, got %q", "static", request.Type)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(struct {
			UUID string `json:"uuid"`
		}{
			UUID: urlsMapUUID[request.Resource],
		})
	}))
	defer ts.Close()

	svc := whiteboard.New(&configs.WhiteboardConfig{
		Endpoint: ts.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := svc.CreateConversionTasks(ctx, urls)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(tasks) != len(urls) {
		t.Errorf("total tasks must match total urls")
	}

	for i, task := range tasks {
		expectedUUID := urlsMapUUID[urls[i]]
		if expectedUUID != task {
			t.Errorf("task UUID does not match")
		}
	}
}

func TestFetchTasksProgress(t *testing.T) {
	t.Parallel()
	tasks := []string{
		"uuid1-finished",
		"uuid2-waiting",
		"uuid3-notfound",
		"uuid4-fail",
		"uuid5-finished",
		"uuid6-finished",
		"uuid7-fail",
		"uuid8-notfound",
		"uuid9-converting",
		"uuid10-converting",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		uuid := strings.TrimPrefix(r.URL.Path, "/v5/services/conversion/tasks/")

		switch {
		case strings.Contains(uuid, "finished"):
			w.Write([]byte(fmt.Sprintf(`
{
	"uuid":%[1]q,
	"status":"Finished",
	"progress":{
		"totalPageSize":2,
		"convertedPageSize":2,
		"convertedPercentage":100,
		"convertedFileList":[
			{"width":1320,"height":1020,"conversionFileUrl":"https://cover.herewhite.com/staticConvert/%[1]s/1.png"},
			{"width":1320,"height":1020,"conversionFileUrl":"https://cover.herewhite.com/staticConvert/%[1]s/2.png"},
			{"width":1320,"height":1020,"conversionFileUrl":"https://cover.herewhite.com/staticConvert/%[1]s/3.png"}
		]
	}
}
`, uuid)))

		case strings.Contains(uuid, "waiting"):
			w.Write([]byte(fmt.Sprintf(`
{
	"uuid":%q,
	"status":"Waiting"
}
`, uuid)))
		case strings.Contains(uuid, "converting"):
			w.Write([]byte(fmt.Sprintf(`
{
	"uuid":%q,
	"status":"Converting"
}
`, uuid)))

		case strings.Contains(uuid, "fail"):
			w.Write([]byte(fmt.Sprintf(`
{
	"uuid":%q,
	"status":"Fail",
	"failedReason": "fail to convert"
}
`, uuid)))

		case strings.Contains(uuid, "notfound"):
			w.Write([]byte(`
{
	"error":{
		"title": "resource not found"
	}
}
`))
		}
	}))

	defer ts.Close()

	svc := whiteboard.New(&configs.WhiteboardConfig{
		Endpoint: ts.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	progresses, err := svc.FetchTasksProgress(ctx, tasks)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(progresses) != len(tasks) {
		t.Errorf("total task progresses must match total tasks")
	}

	for i, progress := range progresses {
		expectedUUID := tasks[i]

		if expectedUUID != progress.UUID {
			t.Errorf("task UUID does not match")
		}
		if strings.Contains(expectedUUID, "waiting") && progress.Progress != nil {
			t.Errorf("expected conversion urls empty")
		}
		if strings.Contains(expectedUUID, "converting") && progress.Progress != nil {
			t.Errorf("expected conversion urls empty")
		}
		if strings.Contains(expectedUUID, "finished") && len(progress.Progress.ConvertedFileList) == 0 {
			t.Errorf("expected conversion urls exists, got empty")
		}
		if strings.Contains(expectedUUID, "fail") && progress.Progress != nil {
			t.Errorf("expected conversion urls empty")
		}
		if strings.Contains(expectedUUID, "notfound") && progress.Error == nil {
			t.Errorf("expected error is not nil")
		}
	}
}
