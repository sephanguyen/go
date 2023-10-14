package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	email_consts "github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/mapper"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mocks "github.com/manabie-com/backend/mock/golibs/kafka"
	mock_commands "github.com/manabie-com/backend/mock/spike/modules/email/application/commands"
	mock_repo "github.com/manabie-com/backend/mock/spike/modules/email/infrastructure/repositories"
	mock_metrics "github.com/manabie-com/backend/mock/spike/modules/email/metrics"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_SendEmail(t *testing.T) {
	t.Parallel()

	createEmailHandler := &mock_commands.MockCreateEmailHandler{}
	mockKafka := mocks.NewKafkaManagement(t)
	mockDB := new(mock_database.Ext)
	mockEmailRepo := mock_repo.MockEmailRepo{}
	mockEmailMetrics := mock_metrics.EmailMetrics{}

	svc := &EmailModifierService{
		EmailCommandHandler: createEmailHandler,
		KafkaMgmt:           mockKafka,
		Env:                 "local",
		DB:                  mockDB,
		EmailRepo:           &mockEmailRepo,
		EmailMetrics:        &mockEmailMetrics,
	}

	emailID := "email-id"

	testCases := []struct {
		Name  string
		Email *spb.SendEmailRequest
		Err   error
		Setup func(ctx context.Context, req *spb.SendEmailRequest)
	}{
		{
			Name: "happy case",
			Email: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					HTML:      "content",
					PlainText: "content",
				},
				Recipients: []string{
					"example-1@manabie.com",
					"example-2@manabie.com",
				},
			},
			Err: nil,
			Setup: func(ctx context.Context, req *spb.SendEmailRequest) {
				createEmailPayload := &commands.CreateEmailPayload{
					Email: mapper.ToEmailDTO(req, email_consts.ManabieDomainEmail),
				}
				email := mapper.ToEmailDTO(req, email_consts.ManabieDomainEmail)
				email.EmailID = emailID
				email.EmailFrom = email_consts.ManabieDomainEmail
				email.Status = spb.EmailStatus_EMAIL_STATUS_QUEUED.String()
				createEmailHandler.On("CreateEmail", ctx, createEmailPayload).Once().Return(email, nil)

				sendEmailEventPayload := mapper.ToSendEmailEventPayload(email)
				sendEmailEventBytesPayload, _ := json.Marshal(sendEmailEventPayload)
				spanName := "PUBLISHER." + constants.EmailSendingTopic
				mockKafka.On("TracedPublishContext", ctx, spanName, constants.EmailSendingTopic, []byte(emailID), sendEmailEventBytesPayload).Once().Return(nil)
				mockEmailMetrics.On("RecordEmailEvents", metrics.EmailQueued, float64(len(req.Recipients))).Once().Return(nil)
			},
		},
		{
			Name: "error publish email event to kafka",
			Email: &spb.SendEmailRequest{
				Subject: "subject",
				Content: &spb.SendEmailRequest_EmailContent{
					HTML:      "content",
					PlainText: "content",
				},
				Recipients: []string{
					"example-1@manabie.com",
					"example-2@manabie.com",
				},
			},
			Err: status.Error(codes.Internal, fmt.Errorf("error when publishing email event: [published failed]").Error()),
			Setup: func(ctx context.Context, req *spb.SendEmailRequest) {
				createEmailPayload := &commands.CreateEmailPayload{
					Email: mapper.ToEmailDTO(req, email_consts.ManabieDomainEmail),
				}

				email := mapper.ToEmailDTO(req, email_consts.ManabieDomainEmail)
				email.EmailID = emailID
				email.EmailFrom = email_consts.ManabieDomainEmail
				email.Status = spb.EmailStatus_EMAIL_STATUS_QUEUED.String()
				createEmailHandler.On("CreateEmail", ctx, createEmailPayload).Once().Return(email, nil)

				sendEmailEventPayload := mapper.ToSendEmailEventPayload(email)
				sendEmailEventBytesPayload, _ := json.Marshal(sendEmailEventPayload)
				spanName := "PUBLISHER." + constants.EmailSendingTopic
				mockKafka.On("TracedPublishContext", ctx, spanName, constants.EmailSendingTopic, []byte(emailID), sendEmailEventBytesPayload).Once().Return(fmt.Errorf("published failed"))

				mockEmailRepo.On("UpdateEmail", ctx, mockDB, emailID, map[string]interface{}{
					"status": spb.EmailStatus_EMAIL_STATUS_QUEUED_FAILED.String(),
				}).Once().Return(nil)
				mockEmailMetrics.On("RecordEmailEvents", metrics.EmailQueued, float64(len(req.Recipients))).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx, tc.Email)

			resp, err := svc.SendEmail(ctx, tc.Email)
			assert.Equal(t, tc.Err, err)
			if tc.Err == nil {
				assert.Equal(t, emailID, resp.EmailId)
			}
		})
	}
}
