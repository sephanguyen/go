package infras

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manabie-com/j4/pkg/instrument"
	j4 "github.com/manabie-com/j4/pkg/runner"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	HasuraQueryNameKey = tag.MustNewKey("hasura_query")
)
var (
	HasuraClientViews = []*view.View{
		ClientCompletedCount,
		ClientRoundtripLatencyDistribution,
	}
	ClientCompletedCount = &view.View{
		Name:        "opencensus.io/http/client/completed_count",
		Measure:     ochttp.ClientRoundtripLatency,
		Aggregation: view.Count(),
		Description: "Count of completed requests, by HTTP method and response status",
		TagKeys:     []tag.Key{ochttp.KeyClientMethod, ochttp.KeyClientStatus, j4.ScenarioKey, HasuraQueryNameKey},
	}
	ClientRoundtripLatencyDistribution = &view.View{
		Name:        "opencensus.io/http/client/roundtrip_latency",
		Measure:     ochttp.ClientRoundtripLatency,
		Aggregation: ochttp.DefaultLatencyDistribution,
		Description: "End-to-end latency, by HTTP method and response status",
		TagKeys:     []tag.Key{ochttp.KeyClientMethod, ochttp.KeyClientStatus, j4.ScenarioKey, HasuraQueryNameKey},
	}
)

func init() {
	instrument.RegisterView(HasuraClientViews...)
}

func (c *Connections) GetHasura(name string) *Hasura {
	h, exist := c.hasura[name]
	if !exist {
		panic(fmt.Sprintf("hasura %s is not created", name))
	}
	return h
}

func (c *Connections) ConnectHasuras(ctx context.Context, cfg *ManabieJ4Config) error {
	if c.hasura == nil {
		c.hasura = map[string]*Hasura{}
	}
	for _, hasuraconf := range cfg.HasuraConfigs {
		h, err := NewHasura(hasuraconf.AdminAddr)
		if err != nil {
			return err
		}
		c.hasura[hasuraconf.Name] = h
	}
	return nil
}

type Hasura struct {
	adminURL string
	cl       *http.Client
}
type NamedQuery struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

// nolint
func NewHasura(adminURl string) (*Hasura, error) {
	cl := &http.Client{
		Transport: hasuraRoundTrip{
			base: &ochttp.Transport{
				Base: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},
	}
	return &Hasura{adminURL: adminURl, cl: cl}, nil
}

type hasuraRoundTrip struct {
	base http.RoundTripper
}
type hasuraQueryNameKey struct{}

func (r hasuraRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	queryname := req.Context().Value(hasuraQueryNameKey{})
	scenarioname, _ := j4.ScenarioNameFromCtx(req.Context())

	ctx, err := tag.New(req.Context(),
		tag.Upsert(HasuraQueryNameKey, queryname.(string)),
		tag.Upsert(j4.ScenarioKey, scenarioname),
	)
	if err != nil {
		return nil, err
	}

	return r.base.RoundTrip(req.WithContext(ctx))
}

func (h *Hasura) QueryRawHasuraV1(ctx context.Context, token, name string, query string, variables map[string]interface{}) ([]byte, error) {
	type GraphQL struct {
		Query     string      `json:"query"`
		Variables interface{} `json:"variables"`
	}

	gr := GraphQL{
		Query:     query,
		Variables: variables,
	}
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(gr)
	if err != nil {
		return nil, err
	}

	reqReader := bytes.NewReader(buf.Bytes())
	ctx = context.WithValue(ctx, hasuraQueryNameKey{}, name)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, h.adminURL+"/v1/graphql", reqReader)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	res, err := h.cl.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		bs, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("%s", string(bs))
	}
	dataRes, err := readHasuraResponse(res.Body)
	if err != nil {
		return nil, err
	}
	return *dataRes, nil
}

func readHasuraResponse(r io.Reader) (*json.RawMessage, error) {
	var out struct {
		Data   *json.RawMessage
		Errors HasuraErrors
	}

	err := json.NewDecoder(r).Decode(&out)

	if err != nil {
		return nil, err
	}

	if len(out.Errors) > 0 {
		return nil, out.Errors
	}
	return out.Data, nil
}

type HasuraErrors []HasuraError

type HasuraError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
	Locations  []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations"`
}

// Error implements error interface.
func (e HasuraErrors) Error() string {
	b := strings.Builder{}
	for _, err := range e {
		b.WriteString(fmt.Sprintf("Message: %s, Locations: %+v", err.Message, err.Locations))
	}
	return b.String()
}
