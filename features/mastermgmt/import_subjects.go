package mastermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (s *suite) checkSubjectImportErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoID, NoName:
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

func (s *suite) checkImportedSubjects(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	allSubjects, err := s.selectAllSubjects(ctx)
	allSubjectsSlice := sliceutils.MapValuesToSlice(allSubjects)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	remains := make([]string, 0, len(stepState.ValidCsvRows))
	for _, row := range stepState.ValidCsvRows {
		rowValues := strings.Split(row, ",")
		id := rowValues[0]
		name := rowValues[1]

		if id != "" {
			v, found := allSubjects[id]
			if !found || v.SubjectID != id {
				remains = append(remains, strings.Join(rowValues, ","))
			}
		} else {
			existName := sliceutils.ContainsFunc(allSubjectsSlice, func(s *domain.Subject) bool {
				return s.Name == name
			})
			if !existName {
				remains = append(remains, strings.Join(rowValues, ","))
			}
		}
	}
	if len(remains) > 0 {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf(
				"not found newly updated subject(s) in DB after updating:\nexpected:\n%s\ngot:\n%v",
				strings.Join(remains, "\r\n"),
				allSubjects,
			)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importsSubjects(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewSubjectServiceClient(s.MasterMgmtConn).
		ImportSubjects(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportSubjectsRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidSubjectsPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	eSubjects, err := s.getExistingSubjects(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	format := "%s,%s"
	timeID := idutil.ULIDNow()
	r1 := fmt.Sprintf(format, eSubjects[0].SubjectID, "updated-name1 "+timeID)
	r2 := fmt.Sprintf(format, eSubjects[1].SubjectID, "updated-name2 "+timeID)
	r3 := fmt.Sprintf(format, "", "updated-name3 "+timeID)

	request := fmt.Sprintf(`subject_id,name
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportSubjectsRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInValidSubjectsPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportSubjectsRequest{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportSubjectsRequest{
				Payload: []byte(`subject_id,name,sequence
				1,gid,name`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 2, got 3"
		}
	case NoID:
		{
			str := "subject_idz,name" + "\n" +
				"id,pID"
			stepState.Request = &mpb.ImportSubjectsRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be subject_id, got subject_idz"
		}
	case NoName:
		{
			str := "subject_id,namez" + "\n" +
				"id,remarks"
			stepState.Request = &mpb.ImportSubjectsRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be name, got namez"
		}
	case WrongLineValues:
		{
			eSubjects, err := s.getExistingSubjects(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			rand.Seed(time.Now().Unix())
			randSubject := eSubjects[0]

			format := "%s,%s"
			// wrong values
			// duplicated name
			r1 := fmt.Sprintf(format, randSubject.SubjectID, randSubject.Name)
			r2 := fmt.Sprintf(format, randSubject.SubjectID, "another")
			// empty name
			r3 := fmt.Sprintf(format, idutil.ULIDNow(), "")

			request := fmt.Sprintf(`subject_id,name
				%s
				%s
				%s`, r1, r2, r3)
			stepState.Request = &mpb.ImportSubjectsRequest{
				Payload: []byte(request),
			}

			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 3",
						Description: "id " + randSubject.SubjectID + " is duplicated",
					},
					{
						Field:       "Row Number: 4",
						Description: "subject name can not be empty",
					},
				},
			}
			stepState.InvalidCsvRows = []string{r1, r2, r3}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllSubjects(ctx context.Context) (map[string]*domain.Subject, error) {
	subjectMap := make(map[string]*domain.Subject)
	stmt :=
		`
		SELECT 
			subject_id,
			name
		FROM
			subject
		`
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query subject")
	}
	defer rows.Close()
	for rows.Next() {
		e := repo.Subject{}
		err := rows.Scan(
			&e.SubjectID,
			&e.Name,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan subject")
		}
		subjectMap[e.SubjectID.String] = e.ToEntity()
	}
	return subjectMap, nil
}

func (s *suite) insertSomeSubjects(ctx context.Context) ([]string, error) {
	size := 4
	ids := make([]string, size)
	for i := 1; i < size; i++ {
		rID := database.Text(idutil.ULIDNow())
		name := database.Text(fmt.Sprintf("Subject %v", rID.String))
		stmt := `INSERT INTO subject
		(subject_id, name, created_at, updated_at)
		VALUES ($1, $2, now(), now())`
		ids = append(ids, rID.String)

		_, err := s.BobDBTrace.Exec(ctx, stmt, rID, name)
		if err != nil {
			return nil, fmt.Errorf("cannot insert subject, err: %s", err)
		}
	}
	return ids, nil
}

func (s *suite) getExistingSubjects(ctx context.Context) ([]*domain.Subject, error) {
	existingSubjects, err := s.selectAllSubjects(ctx)
	if err != nil {
		return nil, err
	}
	if len(existingSubjects) < 3 {
		_, err := s.insertSomeSubjects(ctx)
		if err != nil {
			return nil, err
		}
		existingSubjects, err = s.selectAllSubjects(ctx)
		if err != nil {
			return nil, err
		}
	}
	var (
		es1, es2 *domain.Subject
	)

	for _, v := range existingSubjects {
		if es1 == nil {
			es1 = v
			continue
		}
		if es2 == nil {
			es2 = v
			break
		}
	}
	return []*domain.Subject{
		es1, es2,
	}, nil
}
