package learnosity

import (
	"context"
	"encoding/json"
	"io"
)

// Service represents the name of the API to sign initialisation options for.
type Service uint32

const (
	// ServiceAuthor allows searching and creation of Learnosity powered content from within your content management systems.
	ServiceAuthor Service = iota + 1

	// ServiceItems provides a simple way to embed assessment content from Learnosity's Item Bank into your pages.
	ServiceItems

	// ServiceReports is a cross domain embeddable application that allows content providers to easily embed reports in their pages.
	ServiceReports

	// ServiceData is a back end service that allows consumers to retrieve and store information from within the Learnosity platform.
	ServiceData
)

// Security represents the public and private security keys required to access Learnosity APIs and data.
type Security struct {
	ConsumerKey    string `validate:"required"` // Learnosity consumer key.
	Domain         string `validate:"required"` // Host domain for web server.
	Timestamp      string `validate:"required"` // Timestamp.
	UserID         string `validate:"required"` // Unique identifier.
	ConsumerSecret string `validate:"required"` // Learnosity consumer secret.
}

// RequestString represents the Json stringify format, which is higher priority than Request.
type RequestString string

// Request represents the correct data format to integrate with any of the Learnosity API services.
type Request map[string]any

// Action represents API action types: get, set, update, etc.
type Action string

const (
	ActionNone   Action = ""
	ActionGet    Action = "get"
	ActionSet    Action = "set"
	ActionUpdate Action = "update"
)

// Options are optional parameters.
type Options struct {
	// RequestString represents the Json stringify format, which is higher priority than Request.
	RequestString

	// Request represents the correct data format to integrate with any of the Learnosity API services.
	Request

	// Action represents API action types: get, set, update, etc.
	Action
}

// Option is an interface to apply custom parameters to the Options struct.
type Option interface {
	// Apply modifies the Options struct by setting its fields.
	Apply(*Options)
}

// Apply modifies the Options struct by setting its RequestString field.
func (r RequestString) Apply(opts *Options) {
	opts.RequestString = r
}

// Apply modifies the Options struct by setting its Request field.
func (r Request) Apply(opts *Options) {
	opts.Request = r
}

// Apply modifies the Options struct by setting its Action field.
func (a Action) Apply(opts *Options) {
	opts.Action = a
}

// Init represents the creation and signing of init options for all supported APIs.
type Init interface {
	// Generate used to generate the data necessary to make a request to one of the Learnosity services.
	// If encode is true, the result is a JSON string. Otherwise, it's a map[string]any.
	Generate(encode bool) (any, error)
}

type Meta map[string]any

func (m Meta) Status() bool {
	if ContainsKey(m, "status") {
		return m["status"].(bool)
	}
	return false
}

func (m Meta) Records() int {
	if ContainsKey(m, "records") {
		return int(m["records"].(float64))
	}
	return 0
}

func (m Meta) Next() string {
	if ContainsKey(m, "next") {
		return m["next"].(string)
	}
	return ""
}

// Result represents the response, typical it contains Meta and an array of Data items.
type Result struct {
	Meta Meta            `json:"meta"`
	Data json.RawMessage `json:"data"`
}

// DataAPI represents a back end service that allows consumers to retrieve and store information from within the Learnosity platform.
type DataAPI interface {
	// Request makes a request to Data API. If the data spans multiple pages,
	// then the meta.next property of the response will need to be used to obtain the rest of the data.
	Request(ctx context.Context, http HTTP, endpoint Endpoint, security Security, opts ...Option) (*Result, error)

	// RequestIterator used to iterate over each page of results.
	// This can be useful if the result set is too big to practically fit in memory all at once.
	RequestIterator(ctx context.Context, http HTTP, endpoint Endpoint, security Security, opts ...Option) ([]Result, error)
}

// Method represents HTTP request methods.
type Method string

const (
	MethodGet  Method = "GET"
	MethodPost Method = "POST"
	MethodPut  Method = "PUT"
	MethodDel  Method = "DELETE"
)

// HTTP represents the HTTP client interface.
type HTTP interface {
	// Request makes a HTTP request.
	Request(ctx context.Context, method Method, url string, header map[string]string, body io.Reader, holder any) error
}
