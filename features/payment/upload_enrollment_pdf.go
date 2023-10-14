package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareFileUpload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := []string{
		"package_id",
		"name",
		"package_type",
		"tax_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf(",Package %s,1,,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,2,2022-10-07T00:00:00-07:00,,1,2021-12-07T00:00:00-07:00,2022-10-07T00:00:00-07:00,Remarks,0,0", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",Package %s,2,,2021-12-07,2022-10-07,2,2022-10-07,,1,2021-12-07,2022-10-07,Remarks,0,0", idutil.ULIDNow())
	req := &pb.UploadFileRequest{
		FileName: pb.FileName_ENROLLMENT,
		FileType: pb.FileType_PDF,
		Content: []byte(fmt.Sprintf(`%s 
%s 
%s`, headerText, validRow1, validRow2)),
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) uploadEnrollmentPDF(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewFileServiceClient(s.PaymentConn).UploadFile(contextWithToken(ctx), stepState.Request.(*pb.UploadFileRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getDownloadUrlEnrollmentPDF(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState = StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	res, err := pb.NewFileServiceClient(s.PaymentConn).GetEnrollmentFile(contextWithToken(ctx), &pb.GetEnrollmentFileRequest{})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	previousUrl := stepState.Response.(*pb.UploadFileResponse).DownloadUrl
	if res.DownloadUrl != previousUrl {
		return StepStateToContext(ctx, stepState), fmt.Errorf("get url pdf not equal with previous link, current download url %v, previous download url %v", res.DownloadUrl, previousUrl)
	}
	return StepStateToContext(ctx, stepState), nil
}
