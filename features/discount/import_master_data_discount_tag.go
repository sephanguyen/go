package discount

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

func (s *suite) theInvalidDiscountTagLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportDiscountTagRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportDiscountTagResponse)
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

func (s *suite) theValidDiscountTagLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allDiscountTags, err := s.selectAllDiscountTags(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, row := range stepState.ValidCsvRows {
		rowSplit := strings.Split(row, ",")
		discountTagName := rowSplit[1]
		selectable, err := strconv.ParseBool(rowSplit[2])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		found := false
		for _, e := range allDiscountTags {
			if e.DiscountTagName.String == discountTagName && e.Selectable.Bool == selectable {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountTagValidRequestPayloadWithIncorrectData(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeDiscountTags(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	existingDiscountTags, err := s.selectAllDiscountTags(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",name1 %s,true,false", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",name2 %s,false,false", idutil.ULIDNow())
	validRow3 := fmt.Sprintf(",name3 %s,true,false", idutil.ULIDNow())
	validRow4 := fmt.Sprintf(",name4 %s,true,false", idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",name %s,,false", idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,name %s,,false", existingDiscountTags[1].DiscountTagID.String, idutil.ULIDNow())
	invalidValueRow1 := fmt.Sprintf(",name %s,Selectable,false", idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf("%s,name %s,Selectable,false", existingDiscountTags[2].DiscountTagID.String, idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "empty value row":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(fmt.Sprintf(`discount_tag_id,discount_tag_name,selectable,is_archived
			%s
			%s`, invalidEmptyRow1, invalidEmptyRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(fmt.Sprintf(`discount_tag_id,discount_tag_name,selectable,is_archived
			%s
			%s`, invalidValueRow1, invalidValueRow2)),
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(fmt.Sprintf(`discount_tag_id,discount_tag_name,selectable,is_archived
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingDiscountTag(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.DiscountConn).
		ImportDiscountTag(contextWithToken(ctx), stepState.Request.(*pb.ImportDiscountTagRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountTagValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomeDiscountTags(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	validRow1 := fmt.Sprintf(",name5 %s,true,false", idutil.ULIDNow())
	validRow2 := fmt.Sprintf(",name6 %s,false,false", idutil.ULIDNow())
	stepState.ValidCsvRows = []string{}
	if rowCondition == "all valid rows" {
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(fmt.Sprintf(`discount_tag_id,discount_tag_name,selectable,is_archived
			%s
			%s`, validRow1, validRow2)),
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anDiscountTagInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportDiscountTagRequest{}
	case "header only":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`discount_tag_id,discount_tag_name,selectable,is_archived`),
		}
	case "number of column is not equal 4":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`discount_tag_id,discount_tag_name,is_archived
			,tag_1,true,false
			,tag_2,true,false
			,tag_3,true,false`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`discount_tag_id,discount_tag_name,selectable,is_archived
			,tag_1,true,,false
			,tag_2,true,,false
			,tag_3,true,,false`),
		}
	case "wrong discount_tag_id column name in header":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`Number,discount_tag_name,selectable,is_archived
			,tag_1,true,false
			,tag_2,true,false
			,tag_3,true,false`),
		}
	case "wrong name column name in header":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`discount_tag_id,Naming,selectable,is_archived
			,tag_1,true,false
			,tag_2,true,false
			,tag_3,true,false`),
		}
	case "wrong selectable column name in header":
		stepState.Request = &pb.ImportDiscountTagRequest{
			Payload: []byte(`discount_tag_id,discount_tag_name,Description,is_archived
			,tag_1,true,false
			,tag_2,true,false
			,tag_3,true,false`),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllDiscountTags(ctx context.Context) ([]*entities.DiscountTag, error) {
	allEntities := []*entities.DiscountTag{}
	stmt :=
		`
		SELECT
			discount_tag_id,
			discount_tag_name,
		    selectable,
			is_archived
		FROM
			discount_tag
		`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query discount_tag")
	}

	defer rows.Close()
	for rows.Next() {
		e := &entities.DiscountTag{}
		err := rows.Scan(
			&e.DiscountTagID,
			&e.DiscountTagName,
			&e.Selectable,
			&e.IsArchived,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan discount tag")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *suite) insertSomeDiscountTags(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		randomStr := idutil.ULIDNow()
		name := database.Text("name " + randomStr)
		selectable := database.Bool(rand.Int()%2 == 0)
		stmt := `INSERT INTO discount_tag
		(discount_tag_id, discount_tag_name, selectable, created_at, updated_at, is_archived)
		VALUES ($1, $2, $3, now(), now(), $4)`
		_, err := s.FatimaDBTrace.Exec(ctx, stmt, randomStr, name, selectable, "false")
		if err != nil {
			return fmt.Errorf("cannot insert discount tag, err: %s", err)
		}
	}
	return nil
}

func (s *suite) theImportDiscountTagTransactionIsRolledBack(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	allDiscountTags, err := s.selectAllDiscountTags(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.ValidCsvRows) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	for _, row := range stepState.ValidCsvRows {
		found := false
		rowSplit := strings.Split(row, ",")
		name := rowSplit[1]
		selectable, err := strconv.ParseBool(rowSplit[2])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, e := range allDiscountTags {
			if e.DiscountTagName.String == name && e.Selectable.Bool == selectable {
				found = true
				break
			}
		}
		if found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code().String() != expectedCode {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", expectedCode, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}
