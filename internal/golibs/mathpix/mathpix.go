package mathpix

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/curl"
)

type URL string

var (
	Service URL = "https://api.mathpix.com/v3/text"
)

type Factory interface {
	DetectLatexFromImage(ctx context.Context, content string) ([]Data, error)
}

type FactoryImpl struct {
	HTTP   curl.IHTTP
	AppID  string
	AppKey string
}

type Param struct {
	Src        string          `json:"src"`
	Formats    []string        `json:"formats"`
	DataOption map[string]bool `json:"data_options"`
}

type Result struct {
	Data  []Data `json:"data"`
	Error string `json:"error"`
}

type Data struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m *FactoryImpl) DetectLatexFromImage(ctx context.Context, content string) ([]Data, error) {
	// init data body
	param := &Param{
		Src:        content,
		Formats:    []string{"data"},
		DataOption: map[string]bool{"include_latex": true},
	}
	data, _ := json.Marshal(param)

	// create result
	result := &Result{}

	// make request
	if err := m.HTTP.Request(
		curl.POST,
		string(Service),
		map[string]string{
			"app_id":       m.AppID,
			"app_key":      m.AppKey,
			"Content-Type": "application/json",
		},
		strings.NewReader(string(data)),
		result,
	); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func NewFactory(appID, appKey string, insecureSkipVerify bool) *FactoryImpl {
	return &FactoryImpl{
		HTTP: &curl.HTTP{

			InsecureSkipVerify: insecureSkipVerify,
		},
		AppID:  appID,
		AppKey: appKey,
	}
}
