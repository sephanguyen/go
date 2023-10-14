package payment

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
)

func (s *suite) aNotificationDateValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existingID, err := s.insertNotificationDateAndReturnID(ctx, pb.OrderType_ORDER_TYPE_NEW.String(), 10)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := ",ORDER_TYPE_UPDATE,10,0"
	validRow2 := ",ORDER_TYPE_ENROLLMENT,10,0"
	validRow3 := fmt.Sprintf("%s,ORDER_TYPE_WITHDRAWAL,10,0", existingID)
	validRow4 := ",ORDER_TYPE_LOA,10,0"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	if rowCondition == "all valid rows" {
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(fmt.Sprintf(`notification_date_id,order_type,notification_date,is_archived
        %s
		%s
        %s
		%s`, validRow1, validRow2, validRow3, validRow4)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aNotificationDateValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existingID, err := s.insertNotificationDateAndReturnID(ctx, pb.OrderType_ORDER_TYPE_NEW.String(), 10)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	validRow1 := ",ORDER_TYPE_UPDATE,10,0"
	validRow2 := ",ORDER_TYPE_ENROLLMENT,10,0"
	validRow3 := fmt.Sprintf("%s,ORDER_TYPE_NEW,15,0", existingID)

	invalidEmptyRow1 := ",ORDER_TYPE_CUSTOM_BILLING,,"
	invalidEmptyRow2 := ",,1,"

	invalidValueRow1 := ",ORDER_TYPE_ENROLLMENT,10,3"
	invalidValueRow2 := ",ORDER_TYPE_CUSTOM_BILLING,35,0"

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}

	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(fmt.Sprintf(`notification_date_id,order_type,notification_date,is_archived
        %s
        %s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(fmt.Sprintf(`notification_date_id,order_type,notification_date,is_archived
        %s
        %s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(fmt.Sprintf(`notification_date_id,order_type,notification_date,is_archived
        %s
        %s
        %s
        %s
        %s
        %s
		%s`, validRow1, validRow2, validRow3, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingNotificationDate(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportNotificationDate(contextWithToken(ctx), stepState.Request.(*pb.ImportNotificationDateRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidNotificationDateLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportNotificationDateRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportNotificationDateResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidNotificationDateLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allNotificationDates, err := s.selectAllNotificationDates(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		notificationDateID := rowSplit[0]
		orderType := rowSplit[1]
		notificationDate, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allNotificationDates {
			if notificationDateID == "" {
				if e.OrderType.String == orderType &&
					int(e.NotificationDate.Int) == notificationDate &&
					e.IsArchived.Bool == isArchived &&
					e.CreatedAt.Time.Equal(e.UpdatedAt.Time) {
					found = true
					break
				}
			} else {
				notiDateID := strings.TrimSpace(notificationDateID)
				if e.NotificationDateID.String == notiDateID &&
					e.OrderType.String == orderType &&
					int(e.NotificationDate.Int) == notificationDate &&
					e.IsArchived.Bool == isArchived &&
					e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
					found = true
					break
				}
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theImportNotificationDateTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allNotificationDates, err := s.selectAllNotificationDates(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	for _, row := range stepState.ValidCsvRows {
		found := false

		rowSplit := strings.Split(row, ",")

		notificationDateID := rowSplit[0]
		orderType := rowSplit[1]

		notificationDate, err := strconv.Atoi(strings.TrimSpace(rowSplit[2]))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		isArchived, err := strconv.ParseBool(rowSplit[3])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, e := range allNotificationDates {
			if notificationDateID != "" {
				notiDateID := strings.TrimSpace(notificationDateID)
				if e.NotificationDateID.String == notiDateID &&
					e.OrderType.String == orderType &&
					int(e.NotificationDate.Int) == notificationDate &&
					e.IsArchived.Bool == isArchived &&
					e.CreatedAt.Time.Before(e.UpdatedAt.Time) {
					found = true
					break
				}
			}
		}

		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aNotificationDateInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportNotificationDateRequest{}
	case "header only":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`notification_date_id,order_type,notification_date,is_archived`),
		}
	case "number of column is not equal 4":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`notification_date_id,order_type,notification_date
      ,ORDER_TYPE_CUSTOM_BILLING,10`),
		}
	case "wrong notification_date_id column name in header":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`invalid_notification_date_id,order_type,notification_date,is_archived
      ,ORDER_TYPE_CUSTOM_BILLING,10,1`),
		}
	case "wrong order_type column name in header":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`notification_date_id,invalid_order_type,notification_date,is_archived
      ,ORDER_TYPE_CUSTOM_BILLING,10,1`),
		}
	case "wrong notification_date column name in header":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`notification_date_id,order_type,invalid_notification_date,is_archived
      ,ORDER_TYPE_CUSTOM_BILLING,10,1`),
		}
	case "wrong is_archived column name in header":
		stepState.Request = &pb.ImportNotificationDateRequest{
			Payload: []byte(`notification_date_id,order_type,notification_date,invalid_is_archived
      ,ORDER_TYPE_CUSTOM_BILLING,10,1`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllNotificationDates(ctx context.Context) ([]*entities.NotificationDate, error) {
	allEntities := []*entities.NotificationDate{}
	stmt :=
		`
        SELECT
            notification_date_id,
            order_type,
            notification_date,
            is_archived,
            created_at,
            updated_at
        FROM
            notification_date
        `
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query notification_date")
	}
	defer rows.Close()
	for rows.Next() {
		e := new(entities.NotificationDate)
		err = rows.Scan(
			&e.NotificationDateID,
			&e.OrderType,
			&e.NotificationDate,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan notification date")
		}
		allEntities = append(allEntities, e)
	}

	return allEntities, nil
}

func (s *suite) insertNotificationDateAndReturnID(ctx context.Context, orderType string, notificationDate int) (id string, err error) {
	id = idutil.ULIDNow()
	stmt :=
		`
		INSERT INTO notification_date(
			notification_date_id,
			order_type,
			notification_date,
			is_archived,
			created_at,
			updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
        ON CONFLICT DO NOTHING
		`
	_, err = s.FatimaDBTrace.Exec(ctx, stmt, id, orderType, notificationDate, false)
	if err != nil {
		return
	}
	return
}

func (s *suite) insertSomeNotificationDates(ctx context.Context) (ids []string, err error) {
	ids = make([]string, 0)
	orderTypes := []string{
		pb.OrderType_ORDER_TYPE_LOA.String(),
		pb.OrderType_ORDER_TYPE_RESUME.String(),
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
		pb.OrderType_ORDER_TYPE_GRADUATE.String(),
		pb.OrderType_ORDER_TYPE_NEW.String(),
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
	}
	notificationDate := 10
	for _, orderType := range orderTypes {
		var id string
		id, err = s.insertNotificationDateAndReturnID(ctx, orderType, notificationDate)
		if err != nil {
			return
		}
		ids = append(ids, id)
	}
	return
}
