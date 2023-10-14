package learnosity

import "fmt"

const Version = "v2023.1.LTS"

// Endpoint represents the full url to the endpoint.
type Endpoint string

// Sessions endpoints.
var (
	EndpointDataAPISessionsResponses = Endpoint(fmt.Sprintf("https://data-au.learnosity.com/%s/sessions/responses", Version))
	EndpointDataAPISessionsStatuses  = Endpoint(fmt.Sprintf("https://data-au.learnosity.com/%s/sessions/statuses", Version))
	EndpointDataAPIGetItems          = Endpoint(fmt.Sprintf("https://data-au.learnosity.com/%s/itembank/items", Version))
	EndpointDataAPIGetActivities     = Endpoint(fmt.Sprintf("https://data-au.learnosity.com/%s/itembank/activities", Version))
	// as Learnosity's confirmation, for writing data api we should use url with Oregon region (data-or)
	EndpointDataAPISetItems      = Endpoint(fmt.Sprintf("https://data-or.learnosity.com/%s/itembank/items", Version))
	EndpointDataAPISetFeatures   = Endpoint(fmt.Sprintf("https://data-or.learnosity.com/%s/itembank/features", Version))
	EndpointDataAPISetQuestions  = Endpoint(fmt.Sprintf("https://data-or.learnosity.com/%s/itembank/questions", Version))
	EndpointDataAPISetActivities = Endpoint(fmt.Sprintf("https://data-or.learnosity.com/%s/itembank/activities", Version))
)
