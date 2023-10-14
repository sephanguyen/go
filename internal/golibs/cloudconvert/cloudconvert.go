package cloudconvert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	Host  string
	Token string

	ProjectID       string
	StorageBucket   string
	StorageEndpoint string
	ClientEmail     string
	PrivateKey      string

	Client *http.Client
}

type task struct {
	// import fields
	Operation string `json:"operation,omitempty"`
	URL       string `json:"url,omitempty"`

	// convert fields
	InputFormat  string   `json:"input_format,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"`
	Engine       string   `json:"engine,omitempty"`
	Input        []string `json:"input,omitempty"`
	PixelDensity int      `json:"pixel_density,omitempty"`
	Alpha        bool     `json:"alpha,omitempty"`
	Timeout      int      `json:"timeout,omitempty"`

	// export fields
	ProjectID   string `json:"project_id,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	ClientEmail string `json:"client_email,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	FilePrefix  string `json:"file_prefix,omitempty"`
}

type tasks struct {
	Tasks map[string]task `json:"tasks"`
}

type createJobResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (s *Service) createJob(ctx context.Context, mediaURL string) (*createJobResponse, error) {
	parsedMediaURL, err := url.Parse(mediaURL)
	if err != nil {
		return nil, fmt.Errorf("createJob: url.Parse: %w", err)
	}

	var (
		now = time.Now().UTC().Format("2006-01-02")

		// rnd is used to set the bucket prefix for the converted files
		rnd = idutil.ULIDNow()

		importTaskName  = fmt.Sprintf("import-%s", rnd)
		convertTaskName = fmt.Sprintf("convert-%s", rnd)
		exportTaskName  = fmt.Sprintf("export-%s", rnd)
	)

	jsonReq, err := json.Marshal(tasks{
		Tasks: map[string]task{
			importTaskName: {
				Operation: "import/url",
				URL:       parsedMediaURL.String(), // use parsed url to escape sepcial characters
			},
			convertTaskName: {
				Operation:    "convert",
				InputFormat:  "pdf",
				OutputFormat: "png",
				Engine:       "mupdf",
				Input:        []string{importTaskName},
				PixelDensity: 300,
				Alpha:        false,
				Timeout:      3600,
			},
			exportTaskName: {
				Operation:   "export/google-cloud-storage",
				Input:       []string{convertTaskName},
				ProjectID:   s.ProjectID,
				Bucket:      s.StorageBucket,
				ClientEmail: s.ClientEmail,
				PrivateKey:  s.PrivateKey,
				FilePrefix:  fmt.Sprintf("assignment/%s/images/%s", now, rnd),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v2/jobs", s.Host)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Token))

	request = request.WithContext(ctx)

	resp, err := s.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("createJob: s.Client.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("createJob: got status code: %d, expected: %d", resp.StatusCode, http.StatusCreated)
	}

	var response createJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateConversionTasks creates a conversion task.
func (s *Service) CreateConversionTasks(ctx context.Context, urls []string) ([]string, error) {
	tasks := make([]string, len(urls))

	eg, ctx := errgroup.WithContext(ctx)
	for i, url := range urls {
		i, url := i, url
		eg.Go(func() error {
			resp, err := s.createJob(ctx, url)
			if err != nil {
				return err
			}

			tasks[i] = resp.Data.ID
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// UploadPrefixURL returns the storage bucket url.
func (s *Service) UploadPrefixURL() string {
	return fmt.Sprintf("%s/%s", s.StorageEndpoint, s.StorageBucket)
}
