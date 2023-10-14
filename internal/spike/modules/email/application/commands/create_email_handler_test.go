package commands

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/mapper"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/spike/modules/email/infrastructure/repositories"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := new(mock_database.Ext)
	mockTx := &mock_database.Tx{}
	emailRepo := new(mock_repo.MockEmailRepo)
	emailRecipientRepo := new(mock_repo.MockEmailRecipientRepo)

	handler := CreateEmailHandler{
		DB:                 mockDB,
		EmailRepo:          emailRepo,
		EmailRecipientRepo: emailRecipientRepo,
	}

	req := &spb.SendEmailRequest{
		Subject: "subject",
		Content: &spb.SendEmailRequest_EmailContent{
			PlainText: "content",
			HTML:      "content",
		},
		Recipients: []string{
			"example-1@manabie.com",
			"example-2@manabie.com",
		},
	}

	createEmailPayload := &CreateEmailPayload{
		Email: mapper.ToEmailDTO(req, constants.ManabieDomainEmail),
	}

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)
		emailRepo.On("UpsertEmail", ctx, mockTx, mock.Anything).Return(nil)
		emailRecipientRepo.On("BulkUpsertEmailRecipients", ctx, mockTx, mock.Anything).Return(nil)

		//act
		emailID, err := handler.CreateEmail(ctx, createEmailPayload)

		//assert
		assert.Nil(t, err)
		assert.NotEmpty(t, emailID)
	})

}
