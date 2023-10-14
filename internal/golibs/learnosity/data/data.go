package data

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/learnosity"
	lrni "github.com/manabie-com/backend/internal/golibs/learnosity/init"

	"github.com/pkg/errors"
)

// Client allows consumers to retrieve and store information from within the Learnosity platform.
type Client struct{}

var _ learnosity.DataAPI = (*Client)(nil)

// Request makes a request to DataAPI. Action is get at default.
// If the data spans multiple pages then the meta.next property of the response will need to be used to obtain the rest of the data.
func (c Client) Request(ctx context.Context, http learnosity.HTTP, endpoint learnosity.Endpoint, security learnosity.Security, opts ...learnosity.Option) (*learnosity.Result, error) {
	if endpoint == "" {
		return nil, errors.New(learnosity.ErrNotFoundEndpoint.Error())
	}

	init := lrni.New(learnosity.ServiceData, security, opts...)

	signedRequest, err := init.Generate(false)
	if err != nil {
		return nil, fmt.Errorf("init.Generate: %w", err)
	}
	signedRequestMap := signedRequest.(map[string]any)

	// make request
	formData := url.Values{
		"security": {signedRequestMap["security"].(string)},
		"action":   {signedRequestMap["action"].(string)},
	}
	if signedRequestMap["request"] != nil {
		formData.Set("request", signedRequestMap["request"].(string))
	}

	result := &learnosity.Result{}
	if err = http.Request(
		ctx,
		learnosity.MethodPost,
		string(endpoint),
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		strings.NewReader(formData.Encode()),
		result,
	); err != nil {
		return nil, fmt.Errorf("HTTP.Request: %w", err)
	}

	return result, nil
}

// RequestIterator used to iterate over each page of results.
// this can be useful if the result set is too big to practically fit in memory all at once.
func (c Client) RequestIterator(ctx context.Context, http learnosity.HTTP, endpoint learnosity.Endpoint, security learnosity.Security, opts ...learnosity.Option) ([]learnosity.Result, error) {
	if endpoint == "" {
		return nil, errors.New(learnosity.ErrNotFoundEndpoint.Error())
	}

	// Options struct with default values and applies any given options.
	options := learnosity.Options{
		RequestString: "",
		Request:       nil,
		Action:        learnosity.ActionGet,
	}
	for _, option := range opts {
		option.Apply(&options)
	}

	results := make([]learnosity.Result, 0)
	hasNext := true
	for hasNext {
		result, err := c.Request(ctx, http, endpoint, security, options.RequestString, options.Request, options.Action)
		if err != nil {
			return nil, fmt.Errorf("Request: %w", err)
		}

		meta := result.Meta
		if meta == nil {
			return nil, fmt.Errorf("server returned empty meta: %v", meta)
		}
		if !meta.Status() {
			return nil, fmt.Errorf("server returned unsuccessful status: %v", meta)
		}

		if meta.Records() > 0 {
			results = append(results, *result)

			if options.Request == nil {
				options.Request = make(learnosity.Request, 0)
			}
			options.Request["next"] = meta.Next()
			hasNext = meta.Next() != ""
		} else {
			hasNext = false
		}
	}

	return results, nil
}
