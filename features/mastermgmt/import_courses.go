package mastermgmt

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	NoData            = "no-data"
	WrongColumnCount  = "wrong-column-count"
	NoID              = "no-id-field"
	NoName            = "no-name-field"
	NoRemarks         = "no-remarks-field"
	NoCoursePartnerID = "no-partner_id-field"
	WrongLineValues   = "wrong-line-values"
	NoTypeID          = "no-type_id-field"
	NoIsArchived      = "no-is_archived-field"
)

func (s *suite) importCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).
		ImportCourses(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportCoursesRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	validRows := s.StepState.ValidCsvRows
	expectedCourses := make([]string, 0, len(validRows))
	courseNames := make([]string, 0, len(validRows))
	for _, row := range validRows {
		r := strings.Split(row, ",")
		name := r[1]
		courseTypeID := r[2]
		partnerID := r[3]
		teachingMethod := r[4]
		remarks := r[5]
		schoolID := constants.ManabieSchool
		courseNames = append(courseNames, name)

		expectedCourses = append(expectedCourses, fmt.Sprintf("%s,%s,%s,%s,%s,%d",
			name, courseTypeID, partnerID, remarks, teachingMethod, schoolID))
	}
	var (
		courseName      string
		courseTypeID    pgtype.Text
		coursePartnerID string
		remarks         string
		teachingMethod  string
		resourcePath    int
	)
	query := "SELECT name, course_type_id, course_partner_id, remarks, teaching_method, school_id FROM courses WHERE name = ANY($1) AND deleted_at IS NULL order by updated_at desc"
	rows, err := s.BobDBTrace.Query(ctx, query, courseNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	defer rows.Close()
	respCourse := []string{}
	for rows.Next() {
		if err := rows.Scan(&courseName, &courseTypeID, &coursePartnerID, &remarks, &teachingMethod, &resourcePath); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		csvTeachingMethod := ""
		if teachingMethod == "COURSE_TEACHING_METHOD_GROUP" {
			csvTeachingMethod = "Group"
		} else if teachingMethod == "COURSE_TEACHING_METHOD_INDIVIDUAL" {
			csvTeachingMethod = "Individual"
		}
		course := fmt.Sprintf("%s,%s,%s,%s,%s,%d", courseName, courseTypeID.String, coursePartnerID, remarks, csvTeachingMethod, resourcePath)
		respCourse = append(respCourse, course)
	}

	slices.Sort(expectedCourses)
	slices.Sort(respCourse)
	if equal := slices.Equal(expectedCourses, respCourse); !equal {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courses are not updated properly:\nexpected:%v\ngot:%v", expectedCourses, respCourse)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkImportCourseCSVErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoID, NoName, NoTypeID, NoRemarks:
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

func (s *suite) prepareValidCoursesPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseIDs))))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	nBig2, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseTypeIDs))))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	randCourse := stepState.CourseIDs[nBig.Int64()]
	randCourseType := stepState.CourseTypeIDs[nBig2.Int64()]

	format := "%s,%s,%s,%s,%s,%s"
	timeID := idutil.ULIDNow()
	r1 := fmt.Sprintf(format, randCourse, "course-name "+timeID, randCourseType, timeID+"_pID1", "Group", "updated remarks")
	r2 := fmt.Sprintf(format, idutil.ULIDNow(), "new name-1 "+timeID, randCourseType, timeID+"_pID2", "Individual", "remarks "+timeID)
	// allow empty partner id
	r3 := fmt.Sprintf(format, idutil.ULIDNow(), "new name-2 "+timeID, randCourseType, "", "Group", "")
	r4 := fmt.Sprintf(format, idutil.ULIDNow(), "new name-3 "+timeID, "", timeID+"_pID3", "Group", "")

	request := fmt.Sprintf(`course_id,course_name,course_type_id,course_partner_id,teaching_method,remarks
	%s
	%s
	%s
	%s`, r1, r2, r3, r4)
	stepState.Request = &mpb.ImportCoursesRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2, r3, r4}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInValidCoursesPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportCoursesRequest{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_id
				1,Course 1,0,pid`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 6, got 4"
		}
	case NoID:
		{
			str := "course_idz,course_name,course_type_id,course_partner_id,teaching_method,remarks" + "\n" +
				",name,type_id,pid,Group,remarks"
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be course_id, got course_idz"
		}
	case NoName:
		{
			str := "course_id,course_namez,course_type_id,course_partner_id,teaching_method,remarks" + "\n" +
				"id,name,type_id,pid,Group,remarks"
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be course_name, got course_namez"
		}
	case NoTypeID:
		{
			str := "course_id,course_name,course_typez_id,course_partner_id,teaching_method,remarks" + "\n" +
				"id,z,type_id,pid,Individual,remarks"
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be course_type_id, got course_typez_id"
		}
	case NoCoursePartnerID:
		{
			str := "course_id,course_name,course_type_id,course_partner_idz,teaching_method,remarks" + "\n" +
				"id,z,type_id,pid,Group,remarks"
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 4 should be course_partner_id, got course_partner_idz"
		}
	case NoRemarks:
		{
			str := "course_id,course_name,course_type_id,course_partner_id,teaching_method,remarkz" + "\n" +
				"id,z,type_id,pid,Group,remarks"
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 6 should be remarks, got remarkz"
		}
	case WrongLineValues:
		{
			nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseIDs))))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
			}
			nBig2, err := rand.Int(rand.Reader, big.NewInt(int64(len(stepState.CourseTypeIDs))))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
			}

			randCourse := stepState.CourseIDs[nBig.Int64()]
			randCourseType := stepState.CourseTypeIDs[nBig2.Int64()]

			format := "%s,%s,%s,%s,%s,%s"
			timeID := idutil.ULIDNow()
			// wrong values
			// empty name
			r1 := fmt.Sprintf(format, randCourse, "", randCourseType, timeID+"_pID1", "Group", "updated remarks")
			// empty course_type_id
			r2 := fmt.Sprintf(format, idutil.ULIDNow(), "new name-1 "+timeID, randCourseType, "", "Group", "remarks "+timeID)
			request := fmt.Sprintf(`course_id,course_name,course_type_id,course_partner_id,teaching_method,remarks
				%s
				%s`, r1, r2)
			stepState.Request = &mpb.ImportCoursesRequest{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "name can not be empty",
					},
				},
			}
			stepState.InvalidCsvRows = []string{r1, r2}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
