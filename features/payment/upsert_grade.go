package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/pkg/errors"
)

func (s *suite) anGradeValidRequestPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.NameOfData = fmt.Sprintf("Grade-test-payment %s", idutil.ULIDNow())
	validRow1 := fmt.Sprintf(",%s,1", stepState.NameOfData)
	stepState.Request = &pb.ImportGradeRequest{
		Payload: []byte(fmt.Sprintf(`id,name,is_archived
			%s`, validRow1)),
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingGrade(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewGradeManagementServiceClient(s.BobConn).
		ImportGrade(contextWithToken(ctx), stepState.Request.(*pb.ImportGradeRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentMustSaveConsistentGradeWithBobData(ctx context.Context) (context.Context, error) {
	// sleep to wait for the NATS to deliver the msg
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)
	stmt := `
		SELECT 
			id,
			name,
			is_archived
		FROM
			grade
		WHERE name = $1
		`
	row := s.FatimaDBTrace.QueryRow(
		ctx,
		stmt,
		stepState.NameOfData,
	)
	paymentEntity := &entities.Grade{}
	err := row.Scan(
		&paymentEntity.ID,
		&paymentEntity.Name,
		&paymentEntity.IsArchived,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "rows.Scan grade in payment")
	}

	return StepStateToContext(ctx, stepState), nil
}
