package mastermgmt

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (s *suite) importCourseTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewCourseTypeServiceClient(s.MasterMgmtConn).
		ImportCourseTypes(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportCourseTypesRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedCourseTypes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRows := s.StepState.ValidCsvRows
	expectedCourseTypes := make([]string, 0, len(validRows))
	courseTypeNames := make([]string, 0, len(validRows))
	for _, row := range validRows {
		r := strings.Split(row, ",")
		name := r[1]
		isArchived := r[2]
		remarks := r[3]
		courseTypeNames = append(courseTypeNames, name)
		expectedCourseTypes = append(expectedCourseTypes, fmt.Sprintf("%s,%s,%s",
			name, isArchived, remarks))
	}
	var (
		courseTypeName string
		isArchived     bool
		remarks        string
	)
	query := "SELECT name, is_archived, remarks FROM course_type WHERE name = ANY($1) AND deleted_at IS NULL order by updated_at desc"
	rows, err := s.BobDBTrace.Query(ctx, query, courseTypeNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	defer rows.Close()
	respCourse := []string{}
	for rows.Next() {
		if err := rows.Scan(&courseTypeName, &isArchived, &remarks); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		ct := fmt.Sprintf("%s,%s,%s", courseTypeName, boolToStr(isArchived), remarks)
		respCourse = append(respCourse, ct)
	}

	slices.Sort(respCourse)
	slices.Sort(expectedCourseTypes)
	if equal := slices.Equal(expectedCourseTypes, respCourse); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("course types are not updated properly.\nexpected: %+v\ngot:%+v", expectedCourseTypes, respCourse)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkImportCourseTypeCSVErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoID, NoName, NoRemarks, NoIsArchived:
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
			// Implement error model
			// Validation errors should be returned in error, not in the response.
			return s.compareBadRequest(ctx, err, s.ExpectedErrModel)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidCourseTypesPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseTypeIDs))))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	randCourseType := stepState.CourseTypeIDs[nBig.Int64()]

	format := "%s,%s,%s,%s"
	timeID := idutil.ULIDNow()
	r1 := fmt.Sprintf(format, randCourseType, "course-type-name "+timeID, "0", "updated remarks")
	r2 := fmt.Sprintf(format, idutil.ULIDNow(), "type name-1 "+timeID, "1", "remarks "+timeID)
	r3 := fmt.Sprintf(format, idutil.ULIDNow(), "type name-2 "+timeID, "0", "")

	request := fmt.Sprintf(`course_type_id,course_type_name,is_archived,remarks
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportCourseTypesRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInValidCourseTypesPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportCourseTypesRequest{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(`course_type_id,course_type_name,is_archived
				1,Course 1,0`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 4, got 3"
		}
	case NoID:
		{
			str := "course_type_idz,course_type_name,is_archived,remarks" + "\n" +
				"1,name,0,remarks"
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be course_type_id, got course_type_idz"
		}
	case NoName:
		{
			str := "course_type_id,course_type_namez,is_archived,remarks" + "\n" +
				"id,name,0,remarks"
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be course_type_name, got course_type_namez"
		}
	case NoIsArchived:
		{
			str := "course_type_id,course_type_name,is_archivedz,remarks" + "\n" +
				"id,name,1,remarks"
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be is_archived, got is_archivedz"
		}
	case NoRemarks:
		{
			str := "course_type_id,course_type_name,is_archived,remarksz" + "\n" +
				"id,name,1,remarks"
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 4 should be remarks, got remarksz"
		}
	case WrongLineValues:
		{
			nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseTypeIDs))))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
			}

			randCourseType := stepState.CourseTypeIDs[nBig.Int64()]

			format := "%s,%s,%s,%s"
			timeID := idutil.ULIDNow()
			// wrong values
			// empty name
			r1 := fmt.Sprintf(format, randCourseType, "", "0", "updated remarks")
			// wrong is_archived
			r2 := fmt.Sprintf(format, idutil.ULIDNow(), "new name-2 "+timeID, "bool", "note")

			request := fmt.Sprintf(`course_type_id,course_type_name,is_archived,remarks
				%s
				%s`, r1, r2)
			stepState.Request = &mpb.ImportCourseTypesRequest{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "name can not be empty",
					},
					{
						Field:       "Row Number: 3",
						Description: "bool is not a valid boolean: strconv.ParseBool: parsing \"bool\": invalid syntax",
					},
				},
			}

			stepState.InvalidCsvRows = []string{r1, r2}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
