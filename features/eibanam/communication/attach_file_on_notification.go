package communication

import (
	"context"

	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) generateMediaPDF(name, resource string) *bobPb.Media {
	return &bobPb.Media{
		Name:     name,
		Resource: resource,
		Type:     bobPb.MEDIA_TYPE_PDF,
	}
}

func (s *suite) upsertMedia(ctx context.Context, urls []string) ([]string, error) {
	token := s.getToken(schoolAdmin)

	mediaList := []*bobPb.Media{}
	for _, url := range urls {
		item := s.generateMediaPDF("pdf_file_example", url)
		mediaList = append(mediaList, item)
	}

	req := &bobPb.UpsertMediaRequest{
		Media: mediaList,
	}
	resp, err := bobPb.NewClassClient(s.bobConn).UpsertMedia(contextWithToken(ctx, token), req)
	if err != nil {
		return nil, err
	}

	return resp.MediaIds, nil
}

func (s *suite) schoolAdminSendsNotification(ctx context.Context) (context.Context, error) {
	var err error
	stepState := StepStateFromContext(ctx)
	ctx, err = s.sendNotification(ctx, stepState.notification)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}
