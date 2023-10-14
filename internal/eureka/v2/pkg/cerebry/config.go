package cerebry

import "fmt"

const StagingURL = "https://staging.sparkbackend.cerebry.co"
const PermanentToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjozMzEwNiwidXNlcm5hbWUiOiJ0ZWFjaGVyQGltcy5jb20iLCJleHAiOjE2MzM1MzU1NjIsImVtYWlsIjoidGVhY2hlckBpbXMuY29tIiwib3JpZ19pYXQiOjE2MzM1MjExNjIsImF1ZCI6IlB5dGhvbkFwaSIsImlzcyI6IkNlcmVicnkifQ.-x46jGrmPflE90P43toOaw18gVCyHz6tKeczWKBgo2E" //nolint:gosec

type Config struct {
	BaseURL        string `yaml:"base_url"`
	PermanentToken string `yaml:"permanent_token"`
}

func (c *Config) GetEndpointValue(format RelativeEndpoint, args ...any) string {
	str := fmt.Sprintf(string(format), args...)
	return c.BaseURL + str
}
