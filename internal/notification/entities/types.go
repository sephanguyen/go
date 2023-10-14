package entities

import (
	"encoding/json"
	"strings"
)

type RichText struct {
	Raw         string `json:"raw"`
	RenderedURL string `json:"rendered_url"`
}

func (rc *RichText) GetText() string {
	type block struct {
		Text string `json:"text"`
	}

	type raw struct {
		Blocks []block `json:"blocks"`
	}

	r := raw{}
	err := json.Unmarshal([]byte(rc.Raw), &r)

	if err != nil {
		return ""
	}

	content := []string{}
	for _, block := range r.Blocks {
		// every block is one line
		content = append(content, block.Text)
	}

	return strings.TrimSpace(strings.Join(content, "\n"))
}
