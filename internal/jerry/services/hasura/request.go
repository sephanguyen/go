package hasura

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

type QueryType string

const (
	QueryTypeQuery    QueryType = "query"
	QueryTypeMutation QueryType = "mutation"
)

type Request struct {
	Query     string          `json:"query"`
	Variables json.RawMessage `json:"variables"`

	// Authentication information
	jwt *JWT

	queryType QueryType
	queryName string

	// serviceName is the name of the Hasura service we are working with, e.g. "bob-hasura".
	serviceName string
}

// Key returns the key used to identify cache for this request.
// TODO(@anhpngt): we can consider OperationName instead of Query
// for shorter key.
func (r Request) Key() string {
	return fmt.Sprintf("query:%q|variables:%q|defaultrole:%q|userid:%q|usergroup:%q|resourcepath:%q",
		r.Query, r.Variables, r.jwt.DefaultRole, r.jwt.UserID, r.jwt.UserGroup, r.jwt.ResourcePath)
}

// ID returns a identifying string for this query (usually the query name)
// for metric identification purposes.
func (r Request) ID() string {
	return r.queryName
}

func (r Request) ServiceName() string {
	return r.serviceName
}

// Cachable returns true if Request is cachable.
// Usually, a query of type QueryTypeQuery is cachable.
func (r Request) Cachable() bool {
	return r.queryType == QueryTypeQuery
}

// TagMutators returns the tag.Mutator list for metrics.
func (r Request) TagMutators() []tag.Mutator {
	return []tag.Mutator{
		tag.Upsert(ProxyServiceKey, r.ServiceName()),
		tag.Upsert(QueryNameKey, r.ID()),
	}
}

type Eavesdropper struct {
	svcname string
	logger  *zap.Logger
}

func NewEavesdropper(svcname string, l *zap.Logger) *Eavesdropper {
	return &Eavesdropper{svcname: svcname, logger: l}
}

// EavesdropHTTPRequest extracts Request from r but then reinserts that content
// back into the request's body, so that r is still usable as a http.Request.
func (e *Eavesdropper) EavesdropHTTPRequest(r *http.Request) (*Request, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %s", err)
	}

	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(raw))
	hreq, err := e.parseRequest(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Hasura request: %s", err)
	}
	hjwt, err := e.getHasuraJWTFromHeader(r.Header)
	if err != nil {
		if errors.Is(err, errMissingAuthHeader) {
			e.logger.Error("headers of requests that is missing Auth header", zap.Reflect("headers", r.Header))
		}
		return nil, fmt.Errorf("failed to get JWT info from request: %s", err)
	}
	hreq.jwt = hjwt

	ctx, err := tag.New(r.Context(), tag.Upsert(QueryNameKey, hreq.ID()), tag.Upsert(ProxyServiceKey, hreq.ServiceName()))
	if err != nil {
		return nil, fmt.Errorf("failed to add metric tags to context: %s", err)
	}
	*r = *r.WithContext(ctx)
	return hreq, nil
}

func (e *Eavesdropper) parseRequest(in []byte) (*Request, error) {
	e.logger.Debug("parsing hasura request", zap.ByteString("raw_request", in))
	out := &Request{}
	err := json.Unmarshal(in, out)
	if err != nil {
		return nil, err
	}

	matches := hasuraQueryRe.FindStringSubmatch(out.Query)
	if len(matches) != 3 {
		return nil, fmt.Errorf("failed to parse hasura \"query\" field")
	}

	switch matches[1] {
	case "query":
		out.queryType = QueryTypeQuery
	case "mutation":
		out.queryType = QueryTypeMutation
	default:
		return nil, fmt.Errorf("invalid query type for hasura query: %s", matches[1])
	}

	out.queryName = matches[2]
	if out.queryName == "" {
		e.logger.Warn("empty query name", zap.ByteString("raw_request", in))
	}

	out.serviceName = e.svcname // inject service name for metrics
	return out, err
}

var hasuraQueryRe = regexp.MustCompile("(query|mutation) ([^({ ]*)")

type JWT struct {
	DefaultRole  string `json:"x-hasura-default-role"` // note that http header "X-Hasura-Role" will override this value
	UserID       string `json:"x-hasura-user-id"`
	UserGroup    string `json:"x-hasura-user-group"`
	ResourcePath string `json:"x-hasura-resource-path"`
}

func (jwt JWT) IsValid() bool {
	return jwt.DefaultRole != "" && jwt.UserID != "" && jwt.UserGroup != "" && jwt.ResourcePath != ""
}

var errMissingAuthHeader = errors.New("missing Authorization header")

func (e *Eavesdropper) getHasuraJWTFromHeader(h http.Header) (*JWT, error) {
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		if h.Get("X-Hasura-Admin-Secret") != "" {
			// Hasura admin console is being used instead, return a dummy JWT.
			// This is useful in tests.
			return &JWT{DefaultRole: "x", UserID: "x", UserGroup: "x", ResourcePath: "0"}, nil
		}
		return nil, errMissingAuthHeader
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		e.logger.Error("invalid Authorization header (lacks \"Bearer \" prefix)", zap.String("header", authHeader))
		return nil, fmt.Errorf("invalid auth header")
	}

	tok := authHeader[7:]
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		e.logger.Error("invalid jwt token (does not have 3 parts)", zap.Int("part_count", len(parts)), zap.String("token", tok))
		return nil, fmt.Errorf("invalid jwt token")
	}
	pl, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		e.logger.Error("failed to decode base64 string", zap.String("string", parts[1]))
		return nil, fmt.Errorf("failed to decode base64 string: %s", err)
	}

	var jwtPayload struct {
		JWT JWT `json:"https://hasura.io/jwt/claims"`
	}
	if err := json.Unmarshal(pl, &jwtPayload); err != nil {
		e.logger.Error("failed to unmarshal json", zap.ByteString("jsonstring", pl))
		return nil, fmt.Errorf("failed to unmarshal json: %s", err)
	}
	if !jwtPayload.JWT.IsValid() {
		e.logger.Error("invalid hasura claims in jwt token", zap.Reflect("claims", jwtPayload))
		return nil, fmt.Errorf("invalid hasura claims in jwt token")
	}

	return &jwtPayload.JWT, nil
}
