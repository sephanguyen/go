package speeches

import (
	"context"
	"fmt"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/googleapis/gax-go/v2"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type Text2SpeechBuilder struct {
	client IText2Speech
}

func (builder *Text2SpeechBuilder) NewClient() error {
	ctx, cancle := context.WithTimeout(context.Background(), time.Second*1)
	defer cancle()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("create client error: %w", err)
	}

	builder.client = client
	return nil
}

func (builder *Text2SpeechBuilder) GetClient() IText2Speech {
	return builder.client
}

type IText2Speech interface {
	Close() error
	SynthesizeSpeech(context.Context, *texttospeechpb.SynthesizeSpeechRequest, ...gax.CallOption) (*texttospeechpb.SynthesizeSpeechResponse, error)
}
