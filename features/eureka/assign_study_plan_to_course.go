package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
)

func generateCourseClassEn(courseID string) (*entities.CourseClass, error) {
	var c entities.CourseClass
	database.AllNullEntity(&c)
	rand.Seed(time.Now().UnixNano())
	classID := strconv.Itoa(int(rand.Int31()))
	now := timeutil.Now()
	err := multierr.Combine(
		c.ID.Set(ksuid.New().String()),
		c.CourseID.Set(courseID),
		c.ClassID.Set(classID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	)
	return &c, err
}

func generateCourseStudent(courseID string) (*entities.CourseStudent, error) {
	var c entities.CourseStudent
	database.AllNullEntity(&c)
	rand.Seed(time.Now().UnixNano())
	studentID := idutil.ULIDNow()
	now := timeutil.Now()
	err := multierr.Combine(
		c.ID.Set(ksuid.New().String()),
		c.CourseID.Set(courseID),
		c.StudentID.Set(studentID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	)
	return &c, err
}

func generateClassMemberEn(classID int32) (*entities.ClassStudent, error) {
	var c entities.ClassStudent
	database.AllNullEntity(&c)
	rand.Seed(time.Now().UnixNano())
	studentID := idutil.ULIDNow()
	now := timeutil.Now()
	err := multierr.Combine(
		c.ClassID.Set(classID),
		c.StudentID.Set(studentID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	)
	return &c, err
}

func (s *suite) aValidCourseBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.CourseID) == 0 {
		stepState.CourseID = idutil.ULIDNow()
	}

	ctx, err1 := s.aValidCourseWithStudentInDatabase(ctx, stepState.CourseID)
	if err1 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidCourseWithStudentInDatabase: %w", err1)
	}

	ctx, err2 := s.aValidCourseWithClassInDatabase(ctx, stepState.CourseID)
	if err2 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidCourseWithClassInDatabase: %w", err2)
	}

	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, consta.RoleStudent)
	if ctx, err := s.insertBookIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insertBookIntoBob: %w", err)
	}

	if ctx, err := s.insertChapterIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insertChapterIntoBob: %w", err)
	}

	if ctx, err := s.insertBookChapterIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insertBookChapterIntoBob: %w", err)
	}

	if ctx, err := s.insertTopicIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insertTopicIntoBob: %w", err)
	}

	if ctx, err := s.insertCourseBookWithArgs(ctx, stepState.BookID, stepState.CourseID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insertCourseBook: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createClassInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	cs := &common.Suite{}
	cs.Connections = &common.Connections{
		BobPostgresDB: bobDB,
		BobDB:         bobPgDB,
	}
	cs.StepState = &common.StepState{}
	cs.StepState.FirebaseAddress = firebaseAddr
	cs.StepState.ApplicantID = applicantID

	schoolID, err := s.AdminInsertsSchools(ctx)
	if err != nil {
		return ctx, err
	}

	locationID, _, err := cs.CreateLocationWithDB(ctx, stepState.SchoolID, "center", "", "")
	if err != nil {
		return ctx, fmt.Errorf("cs.CreateLocationWithDB: %v", err)
	}
	stepState.ClassLocationIDs = append(stepState.ClassLocationIDs, locationID)

	ctx, err = cs.UpsertLiveCourse(ctx, stepState.CourseID, []string{}, schoolID)
	if err != nil {
		return ctx, fmt.Errorf("cs.UpsertLiveCourse: %v", err)
	}

	ctx, err = s.aListOfClassInBobDB(ctx, locationID, stepState.CourseID, stepState.ClassIDsString)
	if err != nil {
		return ctx, fmt.Errorf("s.aListOfClassInBobDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AdminInsertsSchools(ctx context.Context) (int32, error) {
	r := &bob_repository.SchoolRepo{}
	city := &bob_entities.City{
		Name:    database.Text("name"),
		Country: database.Text("country"),
	}
	district := &bob_entities.District{
		Name:    database.Text("name"),
		Country: database.Text("country"),
		City:    city,
	}
	school := &bob_entities.School{
		Name:           database.Text("name"),
		Country:        database.Text("country"),
		City:           city,
		District:       district,
		IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present},
		Point:          pgtype.Point{Status: pgtype.Null},
	}
	if err := r.Import(ctx, s.BobDB, []*bob_entities.School{school}); err != nil {
		return 0, err
	}
	return school.ID.Int, nil
}

//nolint:gosec
func (s *suite) aListOfClassInBobDB(ctx context.Context, locationID, courseID string, classIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.Random = fmt.Sprint(rand.Intn(2999) + 1)
	stepState.RequestSentAt = time.Now()

	for _, classID := range classIDs {
		fields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
		query := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
			strings.Join(fields, ","))
		schoolID := golibs.ResourcePathFromCtx(ctx)

		_, err := s.BobDBTrace.Exec(ctx, query, classID, "random-name", courseID, locationID, schoolID, stepState.RequestSentAt, stepState.RequestSentAt)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AValidCourseBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidCourseBackground(ctx)
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aValidCourseWithClassInDatabase(ctx context.Context, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SchoolID == "" {
		stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	}

	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, "USER_GROUP_STUDENT")
	if courseID == "" {
		stepState.CourseID = courseID
	}
	stepState.ClassIDs = []int32{}
	stepState.ClassIDsString = []string{}
	for i := 0; i < 3; i++ {
		courseClass, err := generateCourseClassEn(courseID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		cmd, err := database.Insert(ctx, courseClass, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmd.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course class")
		}
		if classId, err := strconv.Atoi(courseClass.ClassID.String); err != nil {
			return StepStateToContext(ctx, stepState), nil
		} else {
			stepState.ClassIDs = append(stepState.ClassIDs, int32(classId))
			stepState.ClassIDsString = append(stepState.ClassIDsString, courseClass.ClassID.String)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidCourseWithStudentInDatabase(ctx context.Context, courseID string) (_ context.Context, err error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SchoolID == "" {
		stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	}

	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, consta.RoleStudent)
	if courseID == "" {
		stepState.CourseID = courseID
	}

	for i := 0; i < 10; i++ {
		courseStudent, err := generateCourseStudent(courseID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("generateCourseStudent: %w", err)
		}

		cmd, err := database.Insert(ctx, courseStudent, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course student: %w", err)
		}
		if cmd.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course student")
		}
		if _, err := s.aValidUser(ctx, courseStudent.StudentID.String, consta.RoleStudent); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
		}

		stepState.StudentIDs = append(stepState.StudentIDs, courseStudent.StudentID.String)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidClassMemberListInDatabase(ctx context.Context, classIDs []int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var insertClassIDs []int32
	insertClassIDs = append(insertClassIDs, classIDs...)
	if len(insertClassIDs) == 0 {
		insertClassIDs = append(insertClassIDs, stepState.ClassIDs...)
	}
	if len(insertClassIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no class to insert class member")
	}

	for _, classID := range insertClassIDs {
		classMember, err := generateClassMemberEn(classID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("generateClassMemberEn: %w", err)
		}

		cmd, err := database.Insert(ctx, classMember, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert classMember: %w", err)
		}

		if cmd.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course class")
		}
		stepState.StudentIDs = append(stepState.StudentIDs, classMember.StudentID.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidCourseAndStudyPlanBackground(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SchoolID == "" {
		stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	}
	ctx, err1 := s.aValidCourseBackground(ctx)
	if err1 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidCourseBackground:%v", err1)
	}
	ctx, err2 := s.aStudyPlanNameInDb(ctx, idutil.ULIDNow())
	if err2 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aStudyPlanNameInDb:%v", err2)
	}
	ctx, err3 := s.studyPlanAndAssignmentExistsInDb(ctx)
	if err3 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.studyPlanAndAssignmentExistsInDb:%v", err3)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreCourseStudyPlan(ctx context.Context, courseID string, studyPlanID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count(*) FROM course_study_plans csp WHERE csp.course_id = $1 AND csp.study_plan_id = $2`
	var count int
	if err := s.DB.QueryRow(ctx, query, &courseID, &studyPlanID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka store wrong course study plan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreClassStudyPlan(ctx context.Context, classIDs []int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count(*) FROM class_study_plans csp WHERE csp.class_id = ANY($1)`
	var count int
	if err := s.DB.QueryRow(ctx, query, &classIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(classIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka store wrong class study plan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreStudentStudyPlan(ctx context.Context, studentIDs []string, masterStudyPlanID pgtype.Text) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count(*)
		FROM student_study_plans ssp join study_plans sp on ssp.study_plan_id = sp.study_plan_id
		WHERE ssp.student_id = ANY($1) AND ($2::text IS NULL OR sp.master_study_plan_id = $2::text)`
	var count int
	if err := s.DB.QueryRow(ctx, query, &studentIDs, &masterStudyPlanID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(studentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka store wrong student study plan")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) findStudyPLanItems(ctx context.Context, studyPlanID string) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT study_plan_item_id FROM study_plan_items WHERE study_plan_id = $1"
	rows, err := s.DB.Query(ctx, query, &studyPlanID)
	defer rows.Close()
	if err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	var studyPlanItemIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), []string{}, err
		}
		studyPlanItemIDs = append(studyPlanItemIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	return StepStateToContext(ctx, stepState), studyPlanItemIDs, nil
}

func (s *suite) findAssignmentIDsByStudyPlanItem(ctx context.Context, studyPlanItemIDs []string) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT assignment_id FROM assignment_study_plan_items WHERE study_plan_item_id = ANY($1) ORDER BY assignment_id"
	rows, err := s.DB.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	defer rows.Close()
	assignmentIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), []string{}, err
		}
		assignmentIDs = append(assignmentIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	return StepStateToContext(ctx, stepState), assignmentIDs, nil
}

func (s *suite) findLoByStudyPlanItem(ctx context.Context, studyPlanItemIDs []string) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT lo_id FROM lo_study_plan_items WHERE study_plan_item_id = ANY($1) ORDER BY lo_id"
	rows, err := s.DB.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	defer rows.Close()
	loIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), []string{}, err
		}
		loIDs = append(loIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), []string{}, err
	}
	return StepStateToContext(ctx, stepState), loIDs, nil
}

func (s *suite) CopyStudyPlanMustMatchWithOriginal(ctx context.Context, originalStudyPlanID string, copyStudyPLanIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, oStudyPlanItemIDs, err := s.findStudyPLanItems(ctx, originalStudyPlanID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, oAssignments, err := s.findAssignmentIDsByStudyPlanItem(ctx, oStudyPlanItemIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, oLo, err := s.findLoByStudyPlanItem(ctx, oStudyPlanItemIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, copyStudyPlanID := range copyStudyPLanIDs {
		ctx, copystudyPlanItemIDs, err := s.findStudyPLanItems(ctx, copyStudyPlanID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, assignments, err := s.findAssignmentIDsByStudyPlanItem(ctx, copystudyPlanItemIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if !golibs.EqualStringArray(oAssignments, assignments) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("inserted study plan does not match with original")
		}
		ctx, los, err := s.findLoByStudyPlanItem(ctx, copystudyPlanItemIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if !golibs.EqualStringArray(oLo, los) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("inserted study plan does not match with original")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustAssignStudyPlanToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.AssignStudyPlanRequest)
	courseID := req.Data.(*pb.AssignStudyPlanRequest_CourseId).CourseId
	query := "SELECT study_plan_id FROM study_plans WHERE master_study_plan_id= $1"
	rows, err := s.DB.Query(ctx, query, &req.StudyPlanId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	var studyPlanIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyPlanIDs = append(studyPlanIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err1 := s.eurekaMustStoreCourseStudyPlan(ctx, courseID, req.StudyPlanId)
	ctx, err2 := s.eurekaMustStoreStudentStudyPlan(ctx, stepState.StudentIDs, database.Text(req.StudyPlanId))
	ctx, err3 := s.CopyStudyPlanMustMatchWithOriginal(ctx, req.StudyPlanId, studyPlanIDs)

	err = multierr.Combine(err1, err2, err3)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userAssignCourseWithStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.AssignStudyPlanRequest{
		StudyPlanId: stepState.StudyPlanID,
		Data: &pb.AssignStudyPlanRequest_CourseId{
			CourseId: stepState.CourseID,
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).AssignStudyPlan(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
