package salesforce

import (
	"fmt"
	"net/url"
)

const (
	Version   = "v58.0"
	serverURL = "https://manabie4-dev-ed.develop.my.salesforce.com" // TODO: Map domain partner to serverURL "https://<domain>.my.salesforce.com
)

type Endpoint interface {
	GetQueryEndPoint(values url.Values) string
	GetObjectEndPoint(object string) string
}

type EndpointImpl struct{}

func NewEndpoint() Endpoint {
	return &EndpointImpl{}
}

func (e EndpointImpl) GetQueryEndPoint(values url.Values) string {
	endpoint := fmt.Sprintf("/services/data/%s/query?%s", Version, values.Encode())
	return serverURL + endpoint
}

func (e EndpointImpl) GetObjectEndPoint(object string) string {
	endpoint := fmt.Sprintf("/services/data/%s/sobjects/%s", Version, object)
	return serverURL + endpoint
}
