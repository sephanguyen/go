package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/try"
)

type Client struct{}

var _ learnosity.HTTP = (*Client)(nil)

func (c *Client) Request(ctx context.Context, method learnosity.Method, url string, header map[string]string, body io.Reader, holder any) error {
	request, err := http.NewRequestWithContext(ctx, string(method), url, body)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	for key, element := range header {
		request.Header.Set(key, element)
	}

	client := &http.Client{}
	if err := try.Do(func(attempt int) (bool, error) {
		res, err := client.Do(request)
		if err != nil {
			return false, fmt.Errorf("client.Do: %w", err)
		}
		defer res.Body.Close()

		code := res.StatusCode
		if code >= 200 && code <= 299 {
			if holder != nil {
				err := json.NewDecoder(res.Body).Decode(&holder)
				if err != nil {
					return false, fmt.Errorf("json.Decode: %w", err)
				}
			}
			return false, nil
		}

		// If you receive a 429 response, allow the window duration of 5 seconds to elapse before retrying the request.
		if code == 429 {
			time.Sleep(5 * time.Second)
			return attempt < 5, nil
		}

		return attempt < 5, fmt.Errorf("http request failed with status code %d", res.StatusCode)
	}); err != nil {
		return fmt.Errorf("try.Do: %w", err)
	}

	return nil
}
