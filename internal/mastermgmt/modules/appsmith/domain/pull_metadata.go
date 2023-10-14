package domain

type AppsmithResponse struct {
	ResponseMeta AppsmithResponseMeta `json:"responseMeta"`
}

type AppsmithResponseMeta struct {
	Status  int32                 `json:"status"`
	Success bool                  `json:"success"`
	Error   AppsmithResponseError `json:"error,omitempty"`
}

type AppsmithResponseError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
