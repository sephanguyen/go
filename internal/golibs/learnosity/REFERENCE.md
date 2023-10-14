# Learnosity Go-SDK: Reference Guide

## Usage
The following examples show how to use the SDK with the Learnosity APIs.

### Init
Init represents the creation and signing of init options for all supported APIs.

`Init.Generate(encode)` func used to generate the data necessary to make a request to one of the Learnosity services.
If `encode` is *True*, the result is a JSON string. Otherwise, it's a map[string]any.

If the service is Data, `encode` is ignored.

- `Service`: represents the name of the API to sign initialization options for.
- `Security`: represents the public and private security keys required to access Learnosity APIs and data. 
- `RequestString`: represents the JSON stringify format which is higher priority than `Request`.
- `Request`: represents the correct data format to integrate with any of the Learnosity API services.
- `Action`: represents the action type of your request (get, set, update, etc.).

```go
security := learnosity.Security{
	ConsumerKey:    "consumer_key", 
	Domain:         "localhost", 
	Timestamp:      learnosity.FormatUTCTime(time.Now()), 
	UserID:         "user_id", 
	ConsumerSecret: "consumer_secret",
}

requestData := "{\"limit\":5}"

init := lrni.New(learnosity.ServiceItems, security, learnosity.RequestString(requestData))

signedRequest, err := init.Generate(true)
if err != nil {
	log.Fatal("init.Generate: %w", err)
}
```
Note: The `map` type in go is not ordered. So, the reason why we use `RequestString` instead of `Request`
is because the FE or ME can use the correct `signedRequest` to avoid the `Signatures do not match` message error.

### DataAPI

DataAPI represents a back end service that allows consumers to retrieve and store information from within the Learnosity platform.

`DataAPI.Request` func used to make a request to the Data API. Action is `get` at default.

If the data spans multiple pages then the meta.next property of the response will need to be used to obtain the rest of the data.

- `Context`: represents the context of the request.
- `HTTP`: represents the HTTP client to make a request.
- `Endpoint`: represents the full url to the endpoint.
- `Security`: represents the public and private security keys required to access Learnosity APIs and data.
- `Request`: represents the correct data format to integrate with any of the Learnosity API services.
- `Action`: represents the action type of your request (get, set, update, etc.).

Note: `RequestString` parameter is not supported for Data service. Please use `Request` instead.

```go
dataAPI := &lrnd.Client{}
httpClient := &lrnh.Client{}

security := learnosity.Security{
	ConsumerKey:    "consumer_key", 
	Domain:         "localhost", 
	Timestamp:      learnosity.FormatUTCTime(time.Now()), 
	UserID:         "user_id", 
	ConsumerSecret: "consumer_secret",
}

dataRequest := learnosity.Request{
	"activity_id": []string{"activity_id"}, 
	"session_id":  []string{"session_id"}, 
	"user_id":     []string{"user_id"},
}

result, err := dataAPI.Request(ctx, httpClient, learnosity.EndpointDataAPISessionsStatuses, security, dataRequest)
if err != nil {
	log.Fatal("dataAPI.Request: %w", err)
}

dataArr := make([]learnosity.SessionStatus, 0, int(result.Meta["records"].(float64)))
if err := json.Unmarshal(result.Data, &dataArr); err != nil {
	log.Fatal("json.Unmarshal: %w", err)
}
```

Some requests are paginated to the limit passed in the request, or some server-side default. Responses to those requests contain a next parameter in their meta property, which can be placed in the next request to access another page of data.

`DataAPI.RequestIterator` func is the same as `DataAPI.Request` but it supports handle paging.

```go
dataAPI := &lrnd.Client{}
httpClient := &lrnh.Client{}

security := learnosity.Security{
	ConsumerKey:    "consumer_key", 
	Domain:         "localhost", 
	Timestamp:      learnosity.FormatUTCTime(time.Now()), 
	UserID:         "user_id", 
	ConsumerSecret: "consumer_secret",
}

dataRequest := learnosity.Request{
	"activity_id": []string{"activity_id"}, 
	"session_id":  []string{"session_id"}, 
	"user_id":     []string{"user_id"},
}

results, err := dataAPI.RequestIterator(ctx, httpClient, learnosity.EndpointDataAPISessionsResponses, security, dataRequest)
if err != nil {
	log.Fatal("dataAPI.RequestIterator: %w", err)
}

allSessionResponses := make([]learnosity.SessionResponse, 0)
for _, result := range results {
	records := int(result.Meta["records"].(float64))
	
	// Unmarshal the result into a list of SessionResponse. 
	sessionResponses := make([]learnosity.SessionResponse, 0, records)
	if err := json.Unmarshal(result.Data, &sessionResponses); err != nil {
		log.Fatal("json.Unmarshal: %w", err)
	}
	
	allSessionResponses = append(allSessionResponses, sessionResponses)
}
```

### Result Data

As you see above, the `result.Data` is a JSON RawMessage. So, we need to unmarshal it to the correct data type.
To make it easier, we have to define the concrete data type for each endpoint in the `result_data.go` file.

For instance, the `result.Data` of `learnosity.EndpointDataAPISessionsStatuses` is a list of `SessionStatus` data type.

```go
// SessionStatus represents the status returned from the Learnosity Data API.
type SessionStatus struct {
	UserID          string     `json:"user_id"`
	ActivityID      string     `json:"activity_id"`
	NumAttempted    int        `json:"num_attempted"`
	NumQuestions    int        `json:"num_questions"`
	SessionID       string     `json:"session_id"`
	SessionDuration int        `json:"session_duration"`
	Status          string     `json:"status"`
	DtSaved         time.Time  `json:"dt_saved"`
	DtStarted       time.Time  `json:"dt_started"`
	DtCompleted     *time.Time `json:"dt_completed"` // if status is Incomplete, then this field is null
}
```

## Further reading

- Python SDK: https://github.com/Learnosity/learnosity-sdk-python
- Ruby SDK: https://github.com/Learnosity/learnosity-sdk-ruby
- Developer reference docs: https://reference.learnosity.com/
- Demo DataAPI: https://demos.learnosity.com/analytics/data/index.php
