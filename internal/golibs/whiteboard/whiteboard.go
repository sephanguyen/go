package whiteboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Service struct {
	cfg *configs.WhiteboardConfig

	client *http.Client
}

func New(c *configs.WhiteboardConfig) *Service {
	return &Service{
		cfg: c,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

type CreateRoomRequest struct {
	Name     string `json:"name"`
	IsRecord bool   `json:"isRecord"`
	Limit    int    `json:"limit"`
}

type CreateRoomResponse struct {
	UUID string `json:"uuid"`
}

func (s *Service) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v5/rooms", s.cfg.Endpoint)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("token", s.cfg.Token)

	request = request.WithContext(ctx)
	if s.cfg.HttpTracingEnabled {
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace()))
	}

	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("got status code: %d, expected: %d", resp.StatusCode, http.StatusCreated)
	}

	var response CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

type fetchRoomTokenRequest struct {
	Lifespan int    `json:"lifespan"`
	Role     string `json:"role"`
}

func (s *Service) FetchRoomToken(ctx context.Context, roomUUID string) (string, error) {
	jsonReq, err := json.Marshal(fetchRoomTokenRequest{
		Lifespan: int(s.cfg.TokenLifeSpan.Milliseconds()),
		Role:     "admin",
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v5/tokens/rooms/%s", s.cfg.Endpoint, roomUUID)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("token", s.cfg.Token)

	request = request.WithContext(ctx)
	if s.cfg.HttpTracingEnabled {
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace()))
	}

	resp, err := s.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("the HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyString := "null"
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			bodyString = string(bodyBytes)
		}
		return "", fmt.Errorf("got status code: %d, expected: %d respond_body: %s", resp.StatusCode, http.StatusCreated, bodyString)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if body[0] == '"' {
		body = body[1:]
	}
	if i := len(body) - 1; body[i] == '"' {
		body = body[:i]
	}

	return string(body), nil
}

type convertRequest struct {
	Resource string `json:"resource"`
	Type     string `json:"type"`
}

type convertResponse struct {
	UUID string `json:"uuid"`
}

func (s *Service) createConversionTask(ctx context.Context, mediaURL, convertType string) (*convertResponse, error) {
	jsonReq, err := json.Marshal(convertRequest{
		Resource: mediaURL,
		Type:     convertType,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v5/services/conversion/tasks", s.cfg.Endpoint)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("token", s.cfg.Token)

	request = request.WithContext(ctx)

	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("got status code: %d, expected: %d", resp.StatusCode, http.StatusCreated)
	}

	var response convertResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

type CreateConversionTasksResponse struct {
	URL      string
	TaskUUID string
}

// CreateConversionTasks creates a conversion task.
func (s *Service) CreateConversionTasks(ctx context.Context, urls []string) ([]string, error) {
	tasks := make([]string, len(urls))

	eg, ctx := errgroup.WithContext(ctx)
	for i, url := range urls {
		i, url := i, url
		eg.Go(func() error {
			resp, err := s.createConversionTask(ctx, url, "static")
			if err != nil {
				return err
			}

			tasks[i] = resp.UUID
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// {"uuid":"b90def60384111ebafaa5169bd346aaa","type":"static","status":"Waiting"}

// {
//	"uuid":"386267b032d811ebb178f709f5d2b9cc",
//	"status":"Finished",
//	"progress":{
//		"totalPageSize":2,
//		"convertedPageSize":2,
//		"convertedPercentage":100,
//		"convertedFileList":[
//			{"width":1320,"height":1020,"conversionFileUrl":"https://cover.herewhite.com/staticConvert/386267b032d811ebb178f709f5d2b9cc/1.png"},
//			{"width":1320,"height":1020,"conversionFileUrl":"https://cover.herewhite.com/staticConvert/386267b032d811ebb178f709f5d2b9cc/2.png"}
//		]
//	}
// }

type ConvertedFile struct {
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	ConversionFileURL string `json:"conversionFileUrl,omitempty"`
}

type TaskProgress struct {
	TotalPageSize       int             `json:"totalPageSize"`
	ConvertedPercentage float64         `json:"convertedPercentage,omitempty"`
	ConvertedFileList   []ConvertedFile `json:"convertedFileList,omitempty"`
}

type TaskProgressError struct {
	Title string `json:"title,omitempty"`
}

type FetchTaskProgressResponse struct {
	UUID         string             `json:"uuid"`
	Status       string             `json:"status"`
	FailedReason string             `json:"failedReason,omitempty"`
	Progress     *TaskProgress      `json:"progress,omitempty"`
	Error        *TaskProgressError `json:"error,omitempty"`
}

func (s *Service) fetchTaskProgress(ctx context.Context, task string) (*FetchTaskProgressResponse, error) {
	url := fmt.Sprintf("%s/v5/services/conversion/tasks/%s?type=static",
		s.cfg.Endpoint,
		task)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("token", s.cfg.Endpoint)

	request = request.WithContext(ctx)

	resp, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response FetchTaskProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.UUID == "" {
		response.UUID = task
	}

	return &response, nil
}

func (s *Service) FetchTasksProgress(ctx context.Context, tasks []string) ([]*FetchTaskProgressResponse, error) {
	var resp []*FetchTaskProgressResponse
	for _, task := range tasks {
		progress, err := s.fetchTaskProgress(ctx, task)
		if err != nil {
			return nil, err
		}

		resp = append(resp, progress)
	}

	return resp, nil
}

func IsFinishedTask(resp *FetchTaskProgressResponse) bool {
	return resp.Status == "Finished"
}

func trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		ConnectStart: func(network, address string) {
			fmt.Printf("network = %+v, address = %+v, ts = %v\n", network, address, time.Now())
		},
		ConnectDone: func(network, address string, err error) {
			fmt.Printf("network = %+v, address = %+v, err = %+v, ts = %v\n", network, address, err, time.Now())
		},
		GetConn: func(hostPort string) {
			fmt.Printf("hostPort = %+v, ts = %v\n", hostPort, time.Now())
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("connInfo = %+v, ts = %v\n", connInfo, time.Now())
		},
		GotFirstResponseByte: func() {
			fmt.Printf("got first response byte, ts = %+v\n", time.Now())
		},
		WroteRequest: func(requestInfo httptrace.WroteRequestInfo) {
			fmt.Printf("requestInfo = %+v, ts = %v\n", requestInfo, time.Now())
		},
	}
}
