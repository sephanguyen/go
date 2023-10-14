package speeches

import (
	"context"

	"github.com/googleapis/gax-go/v2"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type Text2SpeechClient struct {
}

func (client *Text2SpeechClient) Close() error {
	return nil
}

func (client *Text2SpeechClient) SynthesizeSpeech(ctx context.Context, req *texttospeechpb.SynthesizeSpeechRequest, opts ...gax.CallOption) (*texttospeechpb.SynthesizeSpeechResponse, error) {
	return nil, nil
}
