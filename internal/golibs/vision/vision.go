package vision

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/constants"

	vision "cloud.google.com/go/vision/apiv1"
	gax "github.com/googleapis/gax-go/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	vpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/grpc/codes"
)

type Factory interface {
	DetectTextFromImage(ctx context.Context, content []byte, lang string) (*vpb.TextAnnotation, error)
}

type FactoryImpl struct {
	imageAnnotatorClient *vision.ImageAnnotatorClient
}

func (e *FactoryImpl) DetectTextFromImage(ctx context.Context, content []byte, lang string) (*vpb.TextAnnotation, error) {
	textAnnotation, err := e.imageAnnotatorClient.DetectDocumentText(
		ctx, &vpb.Image{Content: content},
		&vpb.ImageContext{LanguageHints: []string{lang}},
		gax.WithRetry(func() gax.Retryer {
			return gax.OnCodes([]codes.Code{
				codes.Unknown,
				codes.Aborted,
				codes.Unavailable,
				codes.DeadlineExceeded,
			}, gax.Backoff{
				Initial:    constants.DetectTextFromImageRetryInitial,
				Max:        constants.DetectTextFromImageRetryMax,
				Multiplier: constants.DetectTextFromImageRetryMultiplier,
			})
		}),
	)
	if err != nil {
		return nil, err
	}
	return textAnnotation, nil
}

func NewFactory(ctx context.Context, credentials *google.Credentials) (Factory, error) {
	imageAnnotatorClient, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, err
	}
	return &FactoryImpl{
		imageAnnotatorClient: imageAnnotatorClient,
	}, nil
}
