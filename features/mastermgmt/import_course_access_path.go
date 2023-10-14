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
	courseDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	courseRepo "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
	locationDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	locationRepo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	NoCourseID    = "no-course_id-field"
	NotExistingID = "not-exist-values"
	NoLocationID  = "no-location_id-field"
)

func (s *suite) checkCAPImportErrors(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch payloadType {
	case NoData, WrongColumnCount, NoID, NoLocationID, NoCourseID:
		{
			err := stepState.ResponseErr
			if !strings.Contains(err.Error(), stepState.ExpectedError) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong error message with csv format, expected: %s, got: %s", stepState.ExpectedError, err.Error())
			}
		}
	case WrongLineValues, NotExistingID:
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

func (s *suite) checkImportedCAP(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	allCAPs, err := s.selectCAPs(ctx, 2, 200)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	remains := make([]string, 0, len(stepState.ValidCsvRows))
	for _, row := range stepState.ValidCsvRows {
		rowValues := strings.Split(row, ",")
		courseID := rowValues[1]
		locationID := rowValues[2]

		exist := sliceutils.ContainsFunc(allCAPs, func(s courseDomain.CourseAccessPath) bool {
			return s.CourseID == courseID && s.LocationID == locationID
		})
		if !exist {
			remains = append(remains, strings.Join(rowValues, ","))
		}
	}
	if len(remains) > 0 {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf(
				"not found newly updated course access path(s) in DB after updating:\nexpected:\n%s\ngot:\n%v",
				strings.Join(remains, "\r\n"),
				allCAPs,
			)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importCourseAccessPaths(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewCourseAccessPathServiceClient(s.MasterMgmtConn).
		ImportCourseAccessPaths(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportCourseAccessPathsRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareValidCAPPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	exCAPs, err := s.selectCAPs(ctx, 2, 20)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	format := "%s,%s,%s"
	if len(exCAPs) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%s", "please seed more Course access path")
	}
	//nolint
	randNumber := rand.Intn(len(exCAPs))
	randCAP := exCAPs[randNumber]

	// existing course and location
	draftCAPs, err := s.getDraftCourseLocations(ctx, 1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// random existing cap
	r1 := fmt.Sprintf(format, randCAP.ID, randCAP.CourseID, randCAP.LocationID)
	// new cap
	r2 := fmt.Sprintf(format, "", draftCAPs[0].CourseID, draftCAPs[0].LocationID)

	request := fmt.Sprintf(`course_access_path_id,course_id,location_id
	%s
	%s`, r1, r2)
	stepState.Request = &mpb.ImportCourseAccessPathsRequest{
		Payload: []byte(request),
	}
	stepState.ValidCsvRows = []string{r1, r2}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareInValidCAPPayload(ctx context.Context, payloadType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch payloadType {
	case NoData:
		{
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{}
			stepState.ExpectedError = "no data in csv file"
		}
	case WrongColumnCount:
		{
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(`course_access_path_id,course_id,location_id,kien
				1,gid,name,`),
			}
			stepState.ExpectedError = "wrong number of columns, expected 3, got 4"
		}
	case NoID:
		{
			str := "course_access_path_idz,course_id,location_id" + "\n" +
				"id,c_id,l_id"
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 1 should be course_access_path_id, got course_access_path_idz"
		}
	case NoCourseID:
		{
			str := "course_access_path_id,course_idz,location_id" + "\n" +
				"id,c_id,l_id"
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 2 should be course_id, got course_idz"
		}
	case NoLocationID:
		{
			str := "course_access_path_id,course_id,location_idz" + "\n" +
				"id,c_id,l_id"
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(str),
			}
			stepState.ExpectedError = "csv has invalid format, column number 3 should be location_id, got location_idz"
		}
	case NotExistingID:
		{
			exCAP, err := s.getDraftCourseLocations(ctx, 1)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			timeID := idutil.ULIDNow()

			format := "%s,%s,%s"
			r1 := fmt.Sprintf(format, "", "course_id_"+timeID, exCAP[0].LocationID)
			r2 := fmt.Sprintf(format, "", exCAP[0].CourseID, "loc_id_"+timeID)

			request := fmt.Sprintf(`course_access_path_id,course_id,location_id
				%s
				%s`, r1, r2)
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(request),
			}

			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: fmt.Sprintf("course id %s is not exist", "course_id_"+timeID),
					},
					{
						Field:       "Row Number: 3",
						Description: fmt.Sprintf("location id %s is not exist", "loc_id_"+timeID),
					},
				},
			}
			stepState.InvalidCsvRows = []string{r1, r2}
		}
	case WrongLineValues:
		{
			timeID := idutil.ULIDNow()
			format := "%s,%s,%s"
			r1 := fmt.Sprintf(format, "", "course_id_"+timeID, "")
			r2 := fmt.Sprintf(format, "", "", "loc_id_"+timeID)

			request := fmt.Sprintf(`course_access_path_id,course_id,location_id
				%s
				%s`, r1, r2)
			stepState.Request = &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(request),
			}
			stepState.ExpectedError = "data is not valid, please check"
			stepState.ExpectedErrModel = &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: fmt.Sprintf("column %s is required", "location_id"),
					},
					{
						Field:       "Row Number: 3",
						Description: fmt.Sprintf("column %s is required", "course_id"),
					},
				},
			}
			stepState.InvalidCsvRows = []string{r1, r2}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedCourseAccessPaths(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	limit := 5
	draftCAPs, err := s.getDraftCourseLocations(ctx, limit)

	draftCAPsPtr := sliceutils.Map(draftCAPs, func(t courseDomain.CourseAccessPath) *courseDomain.CourseAccessPath {
		return &t
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	expectedRows := [][]string{{
		"course_access_path_id", "course_id", "location_id",
	}}
	for i, c := range draftCAPsPtr {
		e, _ := courseRepo.NewCourseAccessPathFromEntity(c)
		fields, values := e.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		deletedAt := "NULL"
		if i%2 == 0 {
			deletedAt = "now()"
		} else {
			expectedRows = append(expectedRows, []string{
				c.CourseID, c.LocationID,
			})
		}
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT course_access_paths_pk DO UPDATE
		SET updated_at = now(), deleted_at = %s`, e.TableName(), strings.Join(fields, ","), placeHolders, deletedAt)

		_, err := s.BobDBTrace.Exec(ctx, query, values...)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("seedCourseAccessPaths: can not seed course access paths: %w", err)
		}
	}

	stepState.ExpectedCSV = s.getQuotedCSVRows(expectedRows)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedSomeCourses(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < count; i++ {
		timeID := idutil.ULIDNow()
		iStmt := `INSERT INTO courses (
					course_id,
					name,
					grade,
					updated_at,
					created_at)
				VALUES ($1, $2, $3, NOW(), NOW())`
		_, err := s.BobDB.Exec(ctx, iStmt, timeID, timeID+"name", 0)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("seedSomeCourses, err: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

// selectCAPs
// Select course access path
// deleted: 1 only deleted.
// deleted: 0 not deleted.
// deleted: 2, get all.
func (s *suite) selectCAPs(ctx context.Context, deleted int, limit int) ([]courseDomain.CourseAccessPath, error) {
	var caps []courseDomain.CourseAccessPath
	delCond := ""
	if deleted == 0 {
		delCond = "WHERE deleted_at is null"
	} else if deleted == 1 {
		delCond = "WHERE deleted_at is not null"
	}
	stmt :=
		fmt.Sprintf(`
		SELECT 
			id,
			course_id,
			location_id
		FROM
			course_access_paths
		%s
		ORDER BY updated_at DESC
		LIMIT %d
		`, delCond, limit)
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query course_access_paths")
	}
	defer rows.Close()
	for rows.Next() {
		e := courseRepo.CourseAccessPath{}
		err := rows.Scan(
			&e.ID,
			&e.CourseID,
			&e.LocationID,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan course_access_paths")
		}
		caps = append(caps, *e.ToCourseAccessPathEntity())
	}
	return caps, nil
}

func (s *suite) getLocations(ctx context.Context, limit int) ([]locationDomain.Location, error) {
	loc := make([]locationDomain.Location, 0, limit)
	query := fmt.Sprintf(
		`
		SELECT
			l.location_id,
			l.name
		FROM
			public.locations l
		WHERE l.deleted_at is null
		limit %d
		`, limit)
	rows, err := s.BobDBTrace.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, fmt.Errorf("getLocations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := &locationRepo.Location{}
		err := rows.Scan(
			&e.LocationID,
			&e.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("getLocations: rows.Scan %w", err)
		}
		loc = append(loc, *e.ToLocationEntity())
	}
	return loc, nil
}

func (s *suite) getCourses(ctx context.Context, limit int) ([]courseDomain.Course, error) {
	courses := make([]courseDomain.Course, 0, limit)
	stmt := fmt.Sprintf(`
		SELECT
			c.course_id,
			c.name
		FROM
			public.courses c
		WHERE c.deleted_at IS NULL
		order by c.updated_at DESC 
		LIMIT %d
		`, limit)
	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, fmt.Errorf("getCourses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := courseRepo.Course{}
		err := rows.Scan(
			&e.ID,
			&e.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("getCourses: row.Scan %w", err)
		}
		courses = append(courses, *e.ToCourseEntity())
	}
	return courses, nil
}

// getDraftCourseLocations
// get random a couple of course and location in db
func (s *suite) getDraftCourseLocations(ctx context.Context, limit int) ([]courseDomain.CourseAccessPath, error) {
	exLocations, err := s.getLocations(ctx, limit+10)
	if err != nil {
		return nil, fmt.Errorf("getDraftCourseLocations: %s", err)
	}
	if len(exLocations) < limit {
		return nil, fmt.Errorf("getDraftCourseLocations: please seed more location, now we have %d, need %d", len(exLocations), limit)
	}
	exCourses, err := s.getCourses(ctx, limit+10)
	if err != nil {
		return nil, err
	}
	if len(exCourses) < limit {
		return nil, fmt.Errorf("%s", "please seed more courses")
	}

	var caps []courseDomain.CourseAccessPath
	for i := 0; i < limit; i++ {
		//nolint
		randNumber := rand.Intn(len(exLocations))
		randExLocation := exLocations[randNumber]
		//nolint
		randNumber = rand.Intn(len(exCourses))
		randExCourse := exCourses[randNumber]

		caps = append(caps, courseDomain.CourseAccessPath{
			CourseID:   randExCourse.CourseID,
			LocationID: randExLocation.LocationID,
			ID:         idutil.ULIDNow(),
		})
	}
	return caps, nil
}
