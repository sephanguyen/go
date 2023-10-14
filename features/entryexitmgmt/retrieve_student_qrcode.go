package entryexitmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
)

func (s *suite) thisStudentHasQrCodeRecord(ctx context.Context, qrCodeExistence string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if qrCodeExistence != "Existing" {
		return StepStateToContext(ctx, stepState), nil
	}

	var err error
	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	req := &eepb.GenerateBatchQRCodesRequest{
		StudentIds: append(stepState.BatchQRCodeStudentIds, stepState.StudentID),
	}

	res, err := eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).GenerateBatchQRCodes(contextWithToken(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GenerateBatchQRCodes %v", err)
	}
	if res == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response is empty")
	}

	if len(res.Errors) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response has errors: %v", res.Errors)
	}

	if len(res.QrCodes) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("qrcode not created")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentLoginsOnLearnerApp(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	stepState.AuthToken, err = generateValidAuthenticationToken(stepState.StudentID, constant.UserGroupStudent)
	stepState.CurrentUserID = stepState.StudentID
	stepState.CurrentUserGroup = constant.UserGroupStudent
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// set student token
	stepState.AuthToken, err = s.generateExchangeToken(stepState.StudentID, entities.UserGroupStudent, int64(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentIsAtTheMyQRCodeScreen(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &eepb.RetrieveStudentQRCodeRequest{
		StudentId: stepState.StudentID,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRequestedQrCodeWithPayload(ctx context.Context, requestPayload string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if requestPayload == "invalid" {
		stepState.Request.(*eepb.RetrieveStudentQRCodeRequest).StudentId = ""
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).RetrieveStudentQRCode(contextWithToken(ctx), stepState.Request.(*eepb.RetrieveStudentQRCodeRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) StudentQRCodeIsDisplayed(ctx context.Context, displayedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// initialize qr url so it will be reassigned after success response
	switch displayedStatus {
	case "successfully":
		var qrURL string
		if stepState.Response != nil && stepState.ResponseErr == nil {
			qrURL = strings.TrimSpace(stepState.Response.(*eepb.RetrieveStudentQRCodeResponse).QrUrl)
		}

		if qrURL == "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student qr code")
		}
	case "unsuccessfully":
		// expecting error
		if stepState.ResponseErr == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("something went wrong on retrieving qr code")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
