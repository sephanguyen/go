package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const StudyPlanCreate = `ID,Book ID,Book name,Chapter ID,Chapter name,Topic ID,Topic name,Assignment/ LO,Content ID,Name,Available from,Available until,Start time,Due time
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-1,,Learning Objective,Book 1 - Topic 1 - LO 1,LO name 1 - 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-2,,Assignment,assignment-1,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-3,,Assignment,assignment-2,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-3,Book 1-Chapter 3,Topic-ID-7,,Learning Objective,Book 1 - Topic 7 - LO 5,LO name 1 - 5,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-4,Book 1-Chapter 4,Topic-ID-8,,Learning Objective,Book 1 - Topic 8 - LO 6,LO name 1 - 6,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-4,Book 1-Chapter 4,Topic-ID-9,,Learning Objective,Book 1 - Topic 9 - LO 1,LO name 1 - 7,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-4,Book 1-Chapter 4,Topic-ID-11,,Learning Objective,Book 1 - Topic 11 - LO 3,LO name 2 - 2,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-5,Book 1-Chapter 5,Topic-ID-14,,Learning Objective,Book 1 - Topic 14 - LO 2,LO name 3 - 1,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-5,Book 1-Chapter 5,Topic-ID-15,,Learning Objective,Book 1 - Topic 15 - LO 3,LO name 3 - 2,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-5,Book 1-Chapter 5,Topic-ID-16,,Learning Objective,Book 1 - Topic 16 - LO 4,LO name 3 - 3,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-5,Book 1-Chapter 5,Topic-ID-17,,Learning Objective,Book 1 - Topic 17 - LO 5,LO name 3 - 4,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-18,,Learning Objective,Book 1 - Topic 18 - LO 6,LO name 3 - 5,2052-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-19,,Learning Objective,Book 1 - Topic 19 - LO 7,LO name 3 - 6,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-20,,Learning Objective,Book 1 - Topic 20 - LO 8,LO name 3 - 7,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-21,,Learning Objective,Book 1 - Topic 21 - LO 9,LO name 3 - 9,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-6,Book 1-Chapter 6,Topic-ID-22,,Learning Objective,Book 1 - Topic 22 - LO 08,LO name 4 -1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00`

const wrongStudyPlanCreate = `ID,Book ID,Book name,Chapter ID,Chapter name,Topic ID,Topic name,Assignment/ LO,Content ID,Name,Available from,Available until,Start time,Due time
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-1,,Learning Objective,Book 1 - Topic 1 - LO 1,LO name 1 - 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-1,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-2,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-3,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-3,Book 1-Chapter 3,Topic-ID-3,,Learning Objective,Book 1 - Topic 1 - LO 3,LO name 1 - 3,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-4,Book 1-Chapter 4,Topic-ID-4,,Learning Objective,Book 1 - Topic 1 - LO 4,LO name 1 - 4,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00`

const wrongStudyPlanWithBookIDsCreate = `ID,Book ID,Book name,Chapter ID,Chapter name,Topic ID,Topic name,Assignment/ LO,Content ID,Name,Available from,Available until,Start time,Due time
,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-1,,Learning Objective,Book 1 - Topic 1 - LO 1,LO name 1 - 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-1,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-2,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-1,Book 1,B1-Chapter-2,Book 1-Chapter 2,Topic-ID-2,,Assignment,assignment-wrong-3,Assignment name 1,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B1-Chapter-3,Book 2-Chapter 3,Topic-ID-3,,Learning Objective,Book 2 - Topic 1 - LO 3,LO name 1 - 3,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00
,Book-ID-2,Book 2,B1-Chapter-4,Book 2-Chapter 4,Topic-ID-4,,Learning Objective,Book 2 - Topic 1 - LO 4,LO name 1 - 4,2020-08-23T00:00:00-07:00,2020-10-23T11:39:00-07:00,2050-11-23T11:39:00-07:00,2050-11-23T11:39:00-07:00`

const (
	StudyPlanHeader = `ID,Book ID,Book name,Chapter ID,Chapter name,Topic ID,Topic name,Assignment/ LO,Content ID,Name,Available from,Available until,Start time,Due time`
	StudyPlanBody   = `%s,Book-ID-1,Book 1,B1-Chapter-1,Book 1-Chapter 1,Topic-ID-1,,Learning Objective,Book 1 - Topic 1 - LO %d,LO name %d - 1,%s,%s,%s,%s`
)

func (s *suite) UserImportAStudyPlanToCourse(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.userImportAStudyPlanToCourse(ctx, arg)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userImportAStudyPlanToCourse(ctx context.Context, arg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.setFakeClaimToContext(ctx, "1", "USER_GROUP_SCHOOL_ADMIN")
	var studyplan []byte
	// invalid study plan
	switch arg {
	case "invalid":
		studyplan = []byte(wrongStudyPlanCreate)
	case "invalid with books":
		studyplan = []byte(wrongStudyPlanWithBookIDsCreate)
	default:
		if ctx, err := s.insertDataForStudyPlanCreate(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyplan = []byte(StudyPlanCreate)
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), &pb.ImportStudyPlanRequest{
		CourseId: stepState.CourseID,
		Name:     "study-plan-name.csv",
		Mode:     pb.ImportMode_IMPORT_MODE_CREATE,
		Type:     pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
		SchoolId: constants.ManabieSchool,
		Payload:  studyplan,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertDataForStudyPlanCreate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	booksMap := make(map[string]string)
	booksChaptersMap := make(map[string]map[string]bool)
	chaptersMap := make(map[string]string)
	topicsMap := make(map[string]string)
	topicChaptersMap := make(map[string]string)
	lines := strings.Split(StudyPlanCreate, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // ignore header
		}
		parts := strings.Split(line, ",")
		bookID := parts[1]
		bookName := parts[2]
		chapterID := parts[3]
		chapterName := parts[4]
		topicID := parts[5]
		topicName := parts[6]

		booksMap[bookID] = bookName
		chaptersMap[chapterID] = chapterName

		topicsMap[topicID] = topicName
		if _, ok := booksChaptersMap[bookID]; !ok {
			booksChaptersMap[bookID] = make(map[string]bool)
		}
		booksChaptersMap[bookID][chapterID] = true
		topicChaptersMap[topicID] = chapterID
	}

	for bookID, bookName := range booksMap {
		s.insertBookIntoBobWithArgs(ctx, bookID, bookName)
	}

	for chapterID, chapterName := range chaptersMap {
		s.insertChapterIntoBobWithArgs(ctx, chapterID, chapterName)
	}

	for bookID, chaptersMap := range booksChaptersMap {
		for chapterID := range chaptersMap {
			s.insertBookChapterIntoBobWithArgs(ctx, bookID, chapterID)
		}
	}

	for topicID, topicName := range topicsMap {
		s.insertTopicIntoBobWithArgs(ctx, topicID, topicName, topicChaptersMap[topicID])
	}

	s.insertCourseBookIntoBobWithArgs(ctx, stepState.BookID, stepState.CourseID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreAndAssignStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.StudyPlanID = stepState.Response.(*pb.ImportStudyPlanResponse).StudyPlanId
	query := "SELECT study_plan_id FROM study_plans WHERE master_study_plan_id= $1"
	rows, err := s.DB.Query(ctx, query, stepState.StudyPlanID)
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
	ctx, err1 := s.eurekaMustStoreCourseStudyPlan(ctx, stepState.CourseID, stepState.StudyPlanID)
	ctx, err2 := s.eurekaMustStoreStudentStudyPlan(ctx, stepState.StudentIDs, database.Text(stepState.StudyPlanID))
	ctx, err3 := s.CopyStudyPlanMustMatchWithOriginal(ctx, stepState.StudyPlanID, studyPlanIDs)
	err = multierr.Combine(
		err1, err2, err3,
	)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) ValidAssignmentInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.validAssignmentInDb(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) validAssignmentInDb(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SchoolID == "" {
		stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	}
	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())

	if ctx, err := s.insertBookIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.insertChapterIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.insertBookChapterIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.insertTopicIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, ass1 := s.generateAssignment(ctx, "assignment-1", false, false, true)
	ctx, ass2 := s.generateAssignment(ctx, "assignment-2", false, false, true)
	assignments := []*pb.Assignment{
		ass1, ass2,
	}

	req := &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}
	stepState.Request = req
	stepState.Assignments = assignments

	var err error
	stepState.Response, err = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userImportAIndividualStudyPlanToAStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	stepState.Response, err = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), &pb.ImportStudyPlanRequest{
		CourseId:   stepState.CourseID,
		Name:       "individual-study-plan-name.csv",
		Mode:       pb.ImportMode_IMPORT_MODE_CREATE,
		Type:       pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL,
		SchoolId:   constants.ManabieSchool,
		Payload:    []byte(StudyPlanCreate),
		StudentIds: stepState.StudentIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.ResponseErr = err

	stepState.StudyPlanID = stepState.Response.(*pb.ImportStudyPlanResponse).StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustStoreAndAssignStudyPlanForIndividualStudentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.StudyPlanID = stepState.Response.(*pb.ImportStudyPlanResponse).StudyPlanId
	query := `SELECT count(*) FROM study_plan_items WHERE study_plan_id = $1`
	var count int64
	if err := s.DB.QueryRow(ctx, query, stepState.StudyPlanID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != 16 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all study plan item is stored")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userImportAGeneratedStudyPlanToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.initStudyPlanWithGivenTypeAndNumber(ctx, "", 25)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userUpdateCourseStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanFile := StudyPlanHeader
	for i := 0; i < len(stepState.StudyPlanItemIDs); i++ {
		studyPlanFile = fmt.Sprintln(studyPlanFile)
		studyPlanItemID := stepState.StudyPlanItemIDs[i]
		studyPlanFile = studyPlanFile + fmt.Sprintf(
			StudyPlanBody,
			studyPlanItemID,
			i,
			i,
			"",
			"",
			"",
			"",
		)
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), &pb.ImportStudyPlanRequest{
		CourseId:     stepState.CourseID,
		Mode:         pb.ImportMode_IMPORT_MODE_UPDATE,
		Type:         pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
		SchoolId:     constants.ManabieSchool,
		Payload:      []byte(studyPlanFile),
		StudyPlanIds: []string{stepState.StudyPlanID},
		Name:         fmt.Sprintf("updated course study plan %s", stepState.CourseID),
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateCourseStudyPlanWithTimes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	dateBefore := now.AddDate(0, 0, -2)
	dateAfter := now.AddDate(0, 0, 2)

	// 2020-08-23T00:00:00-07:00

	studyPlanFile := StudyPlanHeader
	for i := 0; i < len(stepState.StudyPlanItemIDs); i++ {
		studyPlanFile = fmt.Sprintln(studyPlanFile)
		studyPlanItemID := stepState.StudyPlanItemIDs[i]
		studyPlanFile += fmt.Sprintf(
			StudyPlanBody,
			studyPlanItemID,
			i,
			i,
			dateBefore.Format("2006-01-02T15:04:05-07:00"),
			dateAfter.Format("2006-01-02T15:04:05-07:00"),
			"",
			"",
		)
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), &pb.ImportStudyPlanRequest{
		CourseId:     stepState.CourseID,
		Mode:         pb.ImportMode_IMPORT_MODE_UPDATE,
		Type:         pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE,
		SchoolId:     constants.ManabieSchool,
		Payload:      []byte(studyPlanFile),
		StudyPlanIds: []string{stepState.StudyPlanID},
		Name:         fmt.Sprintf("updated course study plan %s", stepState.CourseID),
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateCourseStudyPlanWithStudyPlanItemsDonotBelongToStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	ctx, err := s.userUpdateCourseStudyPlanWithTimes(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userIndividualStudyPlanMustBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `SELECT student_id FROM course_students WHERE course_id = $1::TEXT`
	rows, err := s.DB.Query(ctx, stmt, &stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	query := `SELECT count(*) FROM study_plan_items spi
	JOIN student_study_plans ssp ON ssp.study_plan_id = spi.study_plan_id
	WHERE copy_study_plan_item_id = ANY($1)
		AND available_from IS NULL
		AND start_date IS NULL
		AND end_date IS NULL
		AND available_to IS NULL
		AND ssp.student_id = $2`

	for _, studentID := range studentIDs {
		var count int64
		err = s.DB.QueryRow(ctx, query, &stepState.StudyPlanItemIDs, &studentID).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if int(count) != len(stepState.StudyPlanItemIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not all study plan item updated")
		}
	}
	ctx, err = s.eurekaMustStoreCourseStudyPlan(ctx, stepState.CourseID, stepState.StudyPlanID)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) getCurrentCourseStudyPlan(ctx context.Context) (context.Context, string, error) {
	stepState := StepStateFromContext(ctx)
	findStudyPlanStmt := `SELECT study_plan_id FROM study_plans sp 
		WHERE sp.course_id = $1::TEXT AND sp.master_study_plan_id IS NULL`
	var studyPlanID string
	err := s.DB.QueryRow(ctx, findStudyPlanStmt, &stepState.CourseID).Scan(&studyPlanID)
	if err != nil {
		return StepStateToContext(ctx, stepState), "", err
	}
	return StepStateToContext(ctx, stepState), studyPlanID, nil
}

func (s *suite) getCurrentStudyPlanItem(ctx context.Context) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)
	findStudyPlanItemsStmt := `SELECT study_plan_item_id FROM study_plan_items spi 
		WHERE study_plan_id=$1::TEXT
			AND spi.deleted_at IS NULL
		ORDER BY spi.display_order ASC`
	rows, err := s.DB.Query(ctx, findStudyPlanItemsStmt, &stepState.StudyPlanID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var studyPlanItemIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		studyPlanItemIDs = append(studyPlanItemIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	return StepStateToContext(ctx, stepState), studyPlanItemIDs, err
}

func (s *suite) studentStudyPlanDisplayOrderMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `SELECT student_id FROM course_students WHERE course_id = $1::TEXT`
	rows, err := s.DB.Query(ctx, stmt, &stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	query := `SELECT count(*) FROM study_plan_items spi
	JOIN student_study_plans ssp ON ssp.study_plan_id = spi.study_plan_id
	JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
	WHERE copy_study_plan_item_id = $1
		AND available_from IS NULL
		AND start_date IS NULL
		AND end_date IS NULL
		AND available_to IS NULL
		AND ssp.student_id = $2
		AND spi.display_order = $3
		AND spi.deleted_at IS NULL`

	for i := range studentIDs {
		var count int64
		for order, copyStudyPlanItemID := range stepState.StudyPlanItemIDs {
			if copyStudyPlanItemID == "" {
				continue
			}
			displayOrder := order
			err = s.DB.QueryRow(ctx, query, &copyStudyPlanItemID, &studentIDs[i], &displayOrder).Scan(&count)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if int(count) != 1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("not all study plan item display order updated. Student: %s don't have study_plan_item_id: %s with display_order: %d",
					studentIDs[i], copyStudyPlanItemID, order)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudentStudyPlanItemMustBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, stepState.StudyPlanID, err = s.getCurrentCourseStudyPlan(ctx)
	ctx, studyPlanItemIDs, err := s.getCurrentStudyPlanItem(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanItemIDs = studyPlanItemIDs

	stmt := `SELECT student_id FROM course_students WHERE course_id = $1`
	rows, err := s.DB.Query(ctx, stmt, &stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	query := `SELECT count(*) FROM study_plan_items spi
	JOIN student_study_plans ssp ON ssp.study_plan_id = spi.study_plan_id
	JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
	WHERE copy_study_plan_item_id = ANY($1::_TEXT)
		AND available_from IS NULL
		AND start_date IS NULL
		AND end_date IS NULL
		AND available_to IS NULL
		AND ssp.student_id = $2::TEXT
		AND spi.deleted_at IS NULL`

	for _, studentID := range studentIDs {
		var count int64
		err = s.DB.QueryRow(ctx, query, &stepState.StudyPlanItemIDs, studentID).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if int(count) != len(stepState.StudyPlanItemIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not all study plan item updated. Expected: %d got %d", len(stepState.StudyPlanItemIDs), count)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) courseStudyPlanAndIndividualStudentStudyPlanMustBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.studentStudyPlanDisplayOrderMustBeUpdated(ctx)
	ctx, err2 := s.allStudentStudyPlanItemMustBeUpdate(ctx)
	ctx, err3 := s.eurekaMustStoreCourseStudyPlan(ctx, stepState.CourseID, stepState.StudyPlanID)
	err := multierr.Combine(
		err1, err2, err3,
	)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userRemoveOneRowFromStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.StudyPlanItemIDs) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no study plan item in study plan")
	}
	stepState.StudyPlanItemIDs = stepState.StudyPlanItemIDs[:len(stepState.StudyPlanItemIDs)-1]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userInsertOneRowToStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, "")
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) individualStudyPlanItemOrderShouldBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT count(*) FROM study_plan_items spi JOIN study_plans sp USING(study_plan_id)
	WHERE spi.study_plan_item_id = $1
		AND available_from IS NULL
		AND start_date IS NULL
		AND end_date IS NULL
		AND available_to IS NULL
		AND spi.deleted_at IS NULL
		AND spi.display_order = $2
		AND sp.study_plan_type='STUDY_PLAN_TYPE_INDIVIDUAL'`
	for i, studyPlanItemID := range stepState.StudyPlanItemIDs {
		var count int64
		if studyPlanItemID == "" {
			continue
		}
		err := s.DB.QueryRow(ctx, query, studyPlanItemID, i).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %s not updated with display order: %d", studyPlanItemID, i)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlanItemSouldBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count(*) FROM study_plan_items spi
	JOIN study_plans sp ON sp.study_plan_id = spi.study_plan_id
	WHERE sp.study_plan_id = $1
		AND available_from IS NULL
		AND start_date IS NULL
		AND end_date IS NULL
		AND available_to IS NULL
		AND spi.deleted_at IS NULL`

	var count int64
	err := s.DB.QueryRow(ctx, query, &stepState.StudyPlanID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != len(stepState.StudyPlanItemIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all study plan item updated. Expected: %d got %d", len(stepState.StudyPlanItemIDs), count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) individualStudyPlanShouldBeUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.individualStudyPlanItemOrderShouldBeUpdate(ctx)
	ctx, err2 := s.allStudyPlanItemSouldBeUpdate(ctx)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) generateUpdatedStudyPlan(ctx context.Context) (context.Context, []byte, int, error) {
	stepState := StepStateFromContext(ctx)
	studyPlanFile := StudyPlanHeader
	for i := 0; i < len(stepState.StudyPlanItemIDs); i++ {
		studyPlanFile = fmt.Sprintln(studyPlanFile)
		studyPlanItemID := stepState.StudyPlanItemIDs[i]
		studyPlanFile = studyPlanFile + fmt.Sprintf(
			StudyPlanBody,
			studyPlanItemID,
			i,
			i,
			"",
			"",
			"",
			"",
		)
	}

	studyPlanContent := []byte(studyPlanFile)
	return StepStateToContext(ctx, stepState), studyPlanContent, len(studyPlanContent), nil
}

func (s *suite) userUpdateIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, buf, n, err := s.generateUpdatedStudyPlan(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), &pb.ImportStudyPlanRequest{
		CourseId:     stepState.CourseID,
		Mode:         pb.ImportMode_IMPORT_MODE_UPDATE,
		Type:         pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL,
		SchoolId:     constants.ManabieSchool,
		Payload:      buf[:n],
		Name:         "Updated individual study plan",
		StudyPlanIds: []string{stepState.StudyPlanID},
		StudentIds:   []string{stepState.StudentIDs[0]},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDownloadAStudentsStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT sp.study_plan_id FROM study_plans sp JOIN student_study_plans ssp
		ON sp.study_plan_id =ssp.study_plan_id WHERE sp.master_study_plan_id = $1::TEXT AND student_id = $2::TEXT`

	var studyPlanID string
	if err := db.QueryRow(ctx, query, &stepState.StudyPlanID, stepState.StudentIDs[0]).Scan(&studyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = studyPlanID
	ctx, studyPlanItems, err := s.getCurrentStudyPlanItem(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanItemIDs = studyPlanItems

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanTask(ctx context.Context, taskID pgtype.Text) (context.Context, pgtype.Text, error) {
	stepState := StepStateFromContext(ctx)
	query := "SELECT status FROM assign_study_plan_tasks WHERE id = $1::TEXT AND deleted_at IS NULL"

	var status pgtype.Text
	if err := s.DB.QueryRow(ctx, query, &taskID).Scan(&status); err != nil {
		return StepStateToContext(ctx, stepState), status, fmt.Errorf("internal error: %w", err)
	}

	return StepStateToContext(ctx, stepState), status, nil
}

func (s *suite) waitForAssignStudyPlanTaskToCompleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.ImportStudyPlanResponse)
	if rsp.TaskId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka must return StepStateToContext(ctx, stepState), task id")
	}

	for {
		ctx, status, err := s.getStudyPlanTask(ctx, database.Text(rsp.TaskId))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if status.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create studyplan")
		}

		if status.String == "STUDY_PLAN_TASK_STATUS_COMPLETED" {
			break
		}
		time.Sleep(time.Microsecond * 10)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userChangeStudyPlanItemOrder(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.StudyPlanItemIDs) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no study plan item in study plan")
	}
	stepState.StudyPlanItemIDs[0], stepState.StudyPlanItemIDs[len(stepState.StudyPlanItemIDs)-1] = stepState.StudyPlanItemIDs[len(stepState.StudyPlanItemIDs)-1], stepState.StudyPlanItemIDs[0]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) initStudyPlanWithGivenTypeAndNumber(ctx context.Context, studyPlanType string, n int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentIds := make([]string, 0)
	typeStudyPlanType := pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE
	if studyPlanType == "individual" {
		typeStudyPlanType = pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL
		studentIds = []string{"1", "2"}
	}

	studyPlanFile := StudyPlanHeader
	for i := 0; i < n; i++ {
		studyPlanFile = fmt.Sprintln(studyPlanFile)
		studyPlanFile += fmt.Sprintf(
			StudyPlanBody,
			"",
			i,
			i,
			"2020-08-23T00:00:00-07:00",
			"2020-10-23T11:39:00-07:00",
			"2050-11-23T11:39:00-07:00",
			"2050-11-23T11:39:00-07:00",
		)
	}
	req := &pb.ImportStudyPlanRequest{
		StudentIds: studentIds,
		CourseId:   stepState.CourseID,
		Mode:       pb.ImportMode_IMPORT_MODE_CREATE,
		Type:       typeStudyPlanType,
		Name:       "generated-study-plan",
		SchoolId:   constants.ManabieSchool,
		Payload:    []byte(studyPlanFile),
	}

	var err error
	stepState.Response, err = pb.NewStudyPlanWriteServiceClient(s.Conn).ImportStudyPlan(contextWithToken(s, ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err = s.waitForAssignStudyPlanTaskToCompleted(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resImportStudyPlan, ok := stepState.Response.(*pb.ImportStudyPlanResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable cast to type *pb.ImportStudyPlanResponse")
	}
	stepState.StudyPlanID = resImportStudyPlan.StudyPlanId
	ctx, stepState.StudyPlanItemIDs, err = s.getCurrentStudyPlanItem(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudyPlanToIndividual(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.initStudyPlanWithGivenTypeAndNumber(ctx, "individual", rand.Intn(10)+10)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) theUserUpdateTheStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.userChangeStudyPlanItemOrder(ctx)
	ctx, err2 := s.userUpdateIndividualStudyPlan(ctx)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) ourSystemHaveToUpdateStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.waitForAssignStudyPlanTaskToCompleted(ctx)
	ctx, err2 := s.individualStudyPlanShouldBeUpdate(ctx)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2)
}

func (s *suite) ourSystemHaveToHandleErrorStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected the response, have to return StepStateToContext(ctx, stepState), error")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) makeStudyPlanItemCompleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `UPDATE study_plan_items
		SET completed_at = NOW()
		WHERE study_plan_item_id = (
			SELECT study_plan_item_id
			FROM study_plan_items
			WHERE study_plan_id = $1
			LIMIT 1
		)`

	if _, err := s.DB.Exec(ctx, query, &stepState.StudyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanItemStillCompleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var studentID string

	query := `SELECT student_id
		FROM student_study_plans
		WHERE study_plan_id = $1::TEXT 
		LIMIT 1`

	if err := db.QueryRow(ctx, query, &stepState.StudyPlanID).Scan(&studentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := try.Do(func(attempt int) (retry bool, err error) {
		resp, err := pb.NewAssignmentReaderServiceClient(conn).RetrieveStudyPlanProgress(contextWithToken(s, ctx), &pb.RetrieveStudyPlanProgressRequest{
			StudyPlanId: stepState.StudyPlanID,
			StudentId:   studentID,
		})
		if err != nil {
			return false, err
		}

		if resp.CompletedAssignments != 1 {
			time.Sleep(2000 * time.Millisecond)
			return attempt < 5, fmt.Errorf("study plan items didn;t complete")
		}
		return false, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func generateLearningObjectiveV1(topicID string) *cpb.LearningObjective {
	num := rand.Int()
	id := fmt.Sprintf("%d", num)
	return &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:           id,
			Name:         "learning",
			Country:      cpb.Country_COUNTRY_VN,
			Grade:        12,
			Subject:      cpb.Subject_SUBJECT_MATHS,
			MasterId:     "",
			DisplayOrder: 1,
			SchoolId:     constants.ManabieSchool,
			CreatedAt:    nil,
			UpdatedAt:    nil,
		},
		TopicId:       topicID,
		Prerequisites: []string{"AL-PH3.1", "AL-PH3.2"},
		StudyGuide:    "https://guides/1/master",
		Video:         "https://videos/1/master",
	}
}

func (s *suite) userTryToUpsertLearningObjectivesUsingAPIV1(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx = contextWithToken(s, ctx)

	stepState.TopicID = "Topic-ID-1"

	var los []*cpb.LearningObjective
	switch arg1 {
	case "valid":
		los = []*cpb.LearningObjective{
			generateLearningObjectiveV1(stepState.TopicID),
			generateLearningObjectiveV1(stepState.TopicID),
		}
		for i := 0; i < len(los); i++ {
			los[i].TopicId = stepState.TopicID
			stepState.LoIDs = append(stepState.LoIDs, los[i].Info.Id)
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unknown: %q", arg1)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(ctx, &pb.UpsertLOsRequest{
		LearningObjectives: los,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlanItemsWereCreatedWithStatusWhenUpsertNewLOs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (bool, error) {
		var total, totalStatusIsNull, totalStatusActive int32

		query := `select 
		count(study_plan_item_id) as total,
		count(study_plan_item_id) filter (where status is null) as with_status_is_null,
		count(study_plan_item_id) filter (where status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::TEXT) as wit_status_is_active
	from study_plan_items spi 
	where spi.content_structure ->> 'lo_id' = any($1::_TEXT);`
		err := db.QueryRow(context.Background(), query, &stepState.LoIDs).Scan(&total, &totalStatusIsNull, &totalStatusActive)
		if err != nil {
			return false, err
		}

		if total != 0 {
			if totalStatusIsNull != 0 {
				return true, fmt.Errorf("expected 0 study plan items have null status, but got %v", totalStatusIsNull)
			}

			if totalStatusActive != total {
				return true, fmt.Errorf("expected %v study plan items have status active, but got %v", total, totalStatusActive)
			}

			return false, nil
		}

		time.Sleep(2 * time.Second)
		return attempt < 5, fmt.Errorf("all study plan items weren't create with status")
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlanItemsWereCreatedWithStatusActiveAfterImportStudyPlanWithTypeCreate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (bool, error) {
		var total, totalStatusActive int32

		query := `select 
		count(study_plan_item_id) as total,
		count(study_plan_item_id) filter (where status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::TEXT) as wit_status_is_active
	from study_plan_items spi 
	where spi.content_structure ->> 'course_id' = $1;`
		err := db.QueryRow(context.Background(), query, &stepState.CourseID).Scan(&total, &totalStatusActive)
		if err != nil {
			return false, err
		}

		if total != 0 {
			if totalStatusActive != total {
				return true, fmt.Errorf("expected %v study plan items have status active, but got %v", total, totalStatusActive)
			}

			return false, nil
		}

		time.Sleep(2 * time.Second)
		return attempt < 5, fmt.Errorf("all study plan items weren't create with status")
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlanItemsWereCreatedWithStatusWhenUpsertNewAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := try.Do(func(attempt int) (bool, error) {
		var total, totalStatusIsNull, totalStatusActive int32

		query := `select 
		count(study_plan_item_id) as total,
		count(study_plan_item_id) filter (where status is null) as with_status_is_null,
		count(study_plan_item_id) filter (where status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'::TEXT) as wit_status_is_active
	from study_plan_items spi 
	where spi.content_structure ->> 'assignment_id' = any($1::_TEXT);`
		err := db.QueryRow(context.Background(), query, &stepState.AssignmentIDs).Scan(&total, &totalStatusIsNull, &totalStatusActive)
		if err != nil {
			return false, err
		}

		if total != 0 {
			if totalStatusIsNull != 0 {
				return true, fmt.Errorf("expected 0 study plan items have null status, but got %v", totalStatusIsNull)
			}

			if totalStatusActive != total {
				return true, fmt.Errorf("expected %v study plan items have status active, but got %v", total, totalStatusActive)
			}

			return false, nil
		}

		time.Sleep(2 * time.Second)
		return attempt < 5, fmt.Errorf("all study plan items weren't create with status")
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userTryToUpsertAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.TopicID = "Topic-ID-1"
	ctx, ass1 := s.generateAssignment(ctx, "", false, false, true)
	ctx, ass2 := s.generateAssignment(ctx, "", false, false, false)
	ctx, ass3 := s.generateAssignment(ctx, "", false, false, false)
	assignments := []*pb.Assignment{
		ass1, ass2, ass3,
	}

	stepState.Assignments = assignments
	for _, assignment := range assignments {
		stepState.AssignmentIDs = append(stepState.AssignmentIDs, assignment.AssignmentId)
	}

	req := &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
