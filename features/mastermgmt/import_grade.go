package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	NoPartnerID = "no-partner_id-field"
	NoSequence  = "no-sequence-field"
)

func (s *suite) checkGradeImportErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoID, NoName, NoPartnerID, NoSequence, NoRemarks:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message with csv format, expected: %s, got: %s", stepState.ExpectedError, err.Error())
			}
		}
	case WrongLineValues:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message for csv line values, expected: %s, got: %s", stepState.ExpectedError, err.Error())
			}
			// Check error model
			return s.compareBadRequest(ctx, err, s.ExpectedErrModel)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkImportedGrades(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	allGrades, err := s.selectAllGrades(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	remains := make([]string, 0, len(stepState.ValidCsvRows))
	for _, row := range stepState.ValidCsvRows {
		rowValues := strings.Split(row, ",")
		partnerID := strings.ToLower(rowValues[1])

		_, found := allGrades[partnerID]
		if !found {
			remains = append(remains, strings.Join(rowValues, ","))
		}
	}
	if len(remains) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found newly updated grade(s) in DB after updating:\r\n %s", strings.Join(remains, "\r\n"))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsGrades(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewGradeServiceClient(s.MasterMgmtConn).
		ImportGrades(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportGradesRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidGradesPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	eGrades, err := s.getExistingGrades(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	rand.Seed(time.Now().Unix())
	//nolint
	randomPID := rand.Int31()

	format := "%s,%s,%s,%s,%s"
	timeID := idutil.ULIDNow()
	r1 := fmt.Sprintf(format, eGrades[0].ID, eGrades[0].PartnerInternalID, "grade-name "+timeID, "200", "updated remarks")
	r2 := fmt.Sprintf(format, eGrades[1].ID, "6000", "grade-name "+timeID, "201", "updated remarks")
	r3 := fmt.Sprintf(format, "", fmt.Sprintf("%d-and-text", randomPID), "grade-name "+timeID, "202", "updated remarks")

	request := fmt.Sprintf(`grade_id,grade_partner_id,name,sequence,remarks
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportGradesRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInValidGradesPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportGradesRequest{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(`grade_id,grade_partner_id,name,sequence
				1,gid,name,1`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 5, got 4"
		}
	case NoID:
		{
			str := "grade_idz,grade_partner_id,name,sequence,remarks" + "\n" +
				"id,pID,name,10,remarks"
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be grade_id, got grade_idz"
		}
	case NoPartnerID:
		{
			str := "grade_id,grade_partner_idz,name,sequence,remarks" + "\n" +
				"id,pID,name,10,remarks"
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be grade_partner_id, got grade_partner_idz"
		}
	case NoName:
		{
			str := "grade_id,grade_partner_id,namez,sequence,remarks" + "\n" +
				"id,pID,name,10,remarks"
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be name, got namez"
		}
	case NoSequence:
		{
			str := "grade_id,grade_partner_id,name,sequencez,remarks" + "\n" +
				"id,pID,name,10,remarks"
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 4 should be sequence, got sequencez"
		}
	case NoRemarks:
		{
			str := "grade_id,grade_partner_id,name,sequence,remarksz" + "\n" +
				"id,pID,name,10,remarks"
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 5 should be remarks, got remarksz"
		}
	case WrongLineValues:
		{
			eGrades, err := s.getExistingGrades(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			gradeIDs := []string{eGrades[0].ID, eGrades[1].ID}

			rand.Seed(time.Now().Unix())
			//nolint
			rIndex := rand.Int31n(int32(len(gradeIDs)))

			randGradeID := gradeIDs[rIndex]

			format := "%s,%s,%s,%s,%s"
			timeID := idutil.ULIDNow()
			// wrong values
			// empty name
			r1 := fmt.Sprintf(format, randGradeID, "1234", "", "2345", "updated remarks")
			// duplicated sequence
			r3 := fmt.Sprintf(format, idutil.ULIDNow(), "1236", "new name-1 "+timeID, "2345", "remarks "+timeID)

			request := fmt.Sprintf(`grade_id,grade_partner_id,name,sequence,remarks
				%s
				%s`, r1, r3)
			stepState.Request = &mpb.ImportGradesRequest{
				Payload: []byte(request),
			}

			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "grade name can not be empty",
					},
					{
						Field:       "Row Number: 3",
						Description: "sequence 2345 is duplicated",
					},
				},
			}
			stepState.InvalidCsvRows = []string{r1, r3}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllGrades(ctx context.Context) (map[string]*domain.Grade, error) {
	gMap := make(map[string]*domain.Grade)
	stmt :=
		`
		SELECT 
			grade_id,
			name,
			partner_internal_id,
			sequence,
			remarks
		FROM
			grade
		`
	rows, err := s.MasterMgmtDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query grade")
	}
	defer rows.Close()
	for rows.Next() {
		e := repo.Grade{}
		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.PartnerInternalID,
			&e.Sequence,
			&e.Remarks,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan grade")
		}
		gMap[e.PartnerInternalID.String] = e.ToGradeEntity()
	}
	return gMap, nil
}

func (s *suite) insertSomeGrades(ctx context.Context) (ids []string, err error) {
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().Unix())

	sSize := 4
	partnerIDs := make([]string, sSize)
	for i := 1; i < sSize; i++ {
		//nolint
		rIndex := rand.Int31n(20)
		rID := database.Text(idutil.ULIDNow())
		name := database.Text(fmt.Sprintf("Grade %v", rID))
		sequence := i
		isArchived := database.Bool(rIndex%2 == 0)
		remarks := fmt.Sprintf("remarks %d", i)
		stmt := `INSERT INTO grade
		(grade_id, name, is_archived, partner_internal_id, sequence, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
		pID := strings.ToLower(idutil.ULIDNow())
		partnerIDs = append(partnerIDs, pID)

		_, err := s.MasterMgmtDBTrace.Exec(ctx, stmt, rID, name, isArchived, database.Varchar(pID), sequence, remarks)
		if err != nil {
			return nil, fmt.Errorf("cannot insert grade, err: %s", err)
		}
	}
	return partnerIDs, nil
}

func (s *suite) getExistingGrades(ctx context.Context) ([]*domain.Grade, error) {
	existingGrades, err := s.selectAllGrades(ctx)
	if err != nil {
		return nil, err
	}
	if len(existingGrades) < 3 {
		_, err := s.insertSomeGrades(ctx)
		if err != nil {
			return nil, err
		}
		existingGrades, err = s.selectAllGrades(ctx)
		if err != nil {
			return nil, err
		}
	}
	var (
		eg1, eg2 *domain.Grade
	)

	for _, v := range existingGrades {
		if eg1 == nil {
			eg1 = v
			continue
		}
		if eg2 == nil {
			eg2 = v
			break
		}
	}
	return []*domain.Grade{
		eg1, eg2,
	}, nil
}
