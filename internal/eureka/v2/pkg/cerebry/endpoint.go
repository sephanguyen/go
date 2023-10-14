package cerebry

// RelativeEndpoint should be relative and not include root URL
// Eg: "/api/v4/partner/user/%s/token/"
type RelativeEndpoint string

var (
	EndpointGenerateUserTokenFmt RelativeEndpoint = "/api/v4/partner/user/%s/token/" //nolint:gosec
)
