package lessonmgmt

import (
	"context"
	crypto_rand "crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	NIL_VALUE string = "nil"
)

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (s *Suite) aListOfLocationTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	listLocationTypes := []struct {
		locationTypeId       string
		name                 string
		parentLocationTypeId string
		archived             bool
		expected             bool
	}{
		// satisfied
		{locationTypeId: "locationtype-id-1", name: "org test", expected: true},
		{locationTypeId: "locationtype-id-2", name: "brand test", parentLocationTypeId: "locationtype-id-1", expected: true},
		{locationTypeId: "locationtype-id-3", name: "area test", parentLocationTypeId: "locationtype-id-1", expected: true},
		{locationTypeId: "locationtype-id-4", name: "center test", parentLocationTypeId: "locationtype-id-2", expected: true},
		{locationTypeId: "locationtype-id-10", name: "center-10", parentLocationTypeId: "locationtype-id-2", expected: true},

		// unsatisfied
		{locationTypeId: "locationtype-id-5", name: "test-5", archived: true},
		{locationTypeId: "locationtype-id-6", name: "test-6", parentLocationTypeId: "locationtype-id-5"},
		{locationTypeId: "locationtype-id-7", name: "test-7", parentLocationTypeId: "locationtype-id-6"},
		{locationTypeId: "locationtype-id-8", name: "test-8", parentLocationTypeId: "locationtype-id-10", archived: true},
		{locationTypeId: "locationtype-id-9", name: "test-9", parentLocationTypeId: "locationtype-id-8"},
	}

	for _, lt := range listLocationTypes {
		stmt := `INSERT INTO location_types (location_type_id,name,parent_location_type_id, is_archived,updated_at,created_at) VALUES($1,$2,$3,$4,now(),now()) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, lt.locationTypeId,
			lt.name,
			NewNullString(lt.parentLocationTypeId),
			lt.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location types with `id:%s`, %v", lt.locationTypeId, err)
		}
		if lt.expected {
			stepState.LocationTypesID = append(stepState.LocationTypesID, lt.locationTypeId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func (s *Suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	nBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(27))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	addedRandom := "-" + strconv.Itoa(int(nBig.Int64()))

	listLocation := []struct {
		locationID        string
		partnerInternalID string
		name              string
		parentLocationID  string
		archived          bool
		expected          bool
		accessPath        string
		locationType      string
	}{ // satisfied
		{locationID: "1" + addedRandom, partnerInternalID: "partner-internal-id-1" + addedRandom, locationType: "locationtype-id-4", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1"})},
		{locationID: "2" + addedRandom, partnerInternalID: "partner-internal-id-2" + addedRandom, locationType: "locationtype-id-5", parentLocationID: "1" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2"})},
		{locationID: "3" + addedRandom, partnerInternalID: "partner-internal-id-3" + addedRandom, locationType: "locationtype-id-6", parentLocationID: "2" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2", "3"})},
		{locationID: "7" + addedRandom, partnerInternalID: "partner-internal-id-7" + addedRandom, locationType: "locationtype-id-7", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7"})},
		// unsatisfied
		{locationID: "4" + addedRandom, partnerInternalID: "partner-internal-id-4" + addedRandom, locationType: "locationtype-id-8", parentLocationID: stepState.LocationID, archived: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4"})},
		{locationID: "5" + addedRandom, partnerInternalID: "partner-internal-id-5" + addedRandom, locationType: "locationtype-id-9", parentLocationID: "4" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "6" + addedRandom, partnerInternalID: "partner-internal-id-6" + addedRandom, locationType: "locationtype-id-1", parentLocationID: "5" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "8" + addedRandom, partnerInternalID: "partner-internal-id-8" + addedRandom, locationType: "locationtype-id-2", parentLocationID: "7" + addedRandom, archived: true, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7", "8"})},
	}

	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,partner_internal_id,name,parent_location_id, is_archived, access_path, location_type) VALUES($1,$2,$3,$4,$5,$6,$7) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, l.locationID, l.partnerInternalID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived, l.accessPath,
			l.locationType)
		if err != nil {
			claims := interceptors.JWTClaimsFromContext(ctx)
			fmt.Println("claims: ", claims.Manabie.UserID, claims.Manabie.ResourcePath)
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) currentStudentAssignedToAboveLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	t, _ := jwt.ParseString(stepState.AuthToken)

	for _, lessonId := range stepState.LessonIDs {
		lessonMember := &entities_bob.LessonMember{}
		database.AllNullEntity(lessonMember)
		if err := multierr.Combine(
			lessonMember.LessonID.Set(lessonId),
			lessonMember.UserID.Set(t.Subject()),
			lessonMember.CreatedAt.Set(now),
			lessonMember.UpdatedAt.Set(now),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if _, err := database.Insert(ctx, lessonMember, s.BobDB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) aListOfLessonsAreExistedInDBOfWithStartTimeAndEndTimeAndLocationID(ctx context.Context, lesson_opt, startTimeString, endTimeString, locationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseID := "course-live-teacher-1"
	courseID2 := "course-live-teacher-1"
	if lesson_opt == "above teacher and belong to multy course" {
		courseID = "course-live-teacher-5"
		courseID2 = "course-live-teacher-6"
	}
	if lesson_opt == "above teacher and belong to single course" {
		courseID = "course-live-teacher-4"
		courseID2 = "course-live-teacher-4"
	}

	courseID += stepState.Random
	courseID2 += stepState.Random
	stepState.CourseIDs = append(stepState.CourseIDs, courseID, courseID2)
	classID := idutil.ULIDNow()

	// create lesson group
	lg := &entities_bob.LessonGroup{}
	database.AllNullEntity(lg)
	lg.MediaIDs = database.TextArray(stepState.MediaIDs)
	lg.CourseID = database.Text(courseID)
	err = (&bob_repo.LessonGroupRepo{}).Create(ctx, s.BobDB, lg)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	for i := 0; i < 20; i++ {
		status := cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()
		if i > 0 {
			status = cpb.LessonSchedulingStatus_name[int32(rand.Intn(4))]
		}
		lesson := &entities_bob.Lesson{}
		lessonID := "bdd_test_retrieve_live_lesson_locations_lesson_id_" + idutil.ULIDNow()
		database.AllNullEntity(lesson)

		err = multierr.Combine(
			lesson.LessonID.Set(lessonID),
			lesson.CourseID.Set(courseID),
			lesson.TeacherID.Set(stepState.CurrentTeacherID),
			lesson.CreatedAt.Set(timeutil.Now()),
			lesson.UpdatedAt.Set(timeutil.Now()),
			lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
			lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
			lesson.StreamLearnerCounter.Set(database.Int4(0)),
			lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
			lesson.StartTime.Set(startDate),
			lesson.EndTime.Set(endDate),
			lesson.LessonGroupID.Set(lg.LessonGroupID),
			lesson.ClassID.Set(classID),
			lesson.CenterID.Set(locationID),
			lesson.SchedulingStatus.Set(database.Text(status)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if err := lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}

		cmdTag, err := database.Insert(ctx, lesson, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}

		if lesson_opt == "above teacher and belong to multy course" {
			sql := `INSERT INTO lessons_courses
				(lesson_id, course_id, created_at)
				VALUES ($1, $2, $4), ($1, $3, $4)`
			_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(courseID),
				database.Text(courseID2),
				database.Timestamptz(time.Now()))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_course, err = %v", err)
			}
		}

		stepState.LessonIDs = append(stepState.LessonIDs, lesson.LessonID.String)

		if err := (&bob_repo.LessonRepo{}).UpsertLessonMembers(ctx, s.BobDB, lesson.LessonID, database.TextArray(stepState.StudentIds)); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) genPresetStudyPlanWeekly(ctx context.Context, startDate, endDate time.Time, lessonID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var topicID string
	if err := s.BobDB.QueryRow(ctx, "SELECT topic_id FROM topics LIMIT 1").Scan(&topicID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var presetStudyPlanID string
	if err := s.BobDB.QueryRow(ctx, "SELECT preset_study_plan_id FROM courses c WHERE c.course_id =$1", courseID).Scan(&presetStudyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	presetStudyPlanWeekly := &entities_bob.PresetStudyPlanWeekly{}
	database.AllNullEntity(presetStudyPlanWeekly)
	presetStudyPlanWeeklyID := "bdd_test_preset_study_plan_weekly+id_" + idutil.ULIDNow()
	week := rand.Intn(10000)
	err := multierr.Combine(
		presetStudyPlanWeekly.ID.Set(presetStudyPlanWeeklyID),
		presetStudyPlanWeekly.StartDate.Set(startDate),
		presetStudyPlanWeekly.EndDate.Set(endDate),
		presetStudyPlanWeekly.PresetStudyPlanID.Set(presetStudyPlanID),
		presetStudyPlanWeekly.TopicID.Set(topicID),
		presetStudyPlanWeekly.Week.Set(week),
		presetStudyPlanWeekly.CreatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.UpdatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.LessonID.Set(lessonID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	cmdTag, err := database.Insert(ctx, presetStudyPlanWeekly, s.BobDB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert preset study plan weekly")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) userRetrieveLiveLessonWithStartTimeAndEndTimeAndLocationID(ctx context.Context, userRole, startTimeString, endTimeString, locationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = stepState.CurrentTeacherID
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseIDs := stepState.CourseIDs
	req := &bpb.RetrieveLiveLessonByLocationsRequest{
		Pagination: &bpb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From:        &timestamppb.Timestamp{Seconds: startDate.Unix()},
		To:          &timestamppb.Timestamp{Seconds: endDate.Unix()},
		LocationIds: []string{locationID},
		CourseIds:   courseIDs,
	}
	stepState.Request = req

	if userRole == "student" {
		stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.BobConn).RetrieveLiveLessonByLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	}
	if userRole == "teacher" {
		token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.BobConn).RetrieveLiveLessonByLocations(helper.GRPCContext(ctx, "token", token), req)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userRetrieveLiveLessonByCourseWithStartTimeAndEndTimeAndLocationID(ctx context.Context, userRole, courseID, startTimeString, endTimeString, locationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = stepState.CurrentTeacherID
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseID += stepState.Random

	req := &bpb.RetrieveLiveLessonByLocationsRequest{
		Pagination: &bpb.Pagination{
			Limit: 100,
			Page:  1,
		},
		CourseIds:   []string{courseID},
		From:        &timestamppb.Timestamp{Seconds: startDate.Unix()},
		To:          &timestamppb.Timestamp{Seconds: endDate.Unix()},
		LocationIds: []string{locationID},
	}
	stepState.Request = req
	if userRole == "student" {
		stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.BobConn).RetrieveLiveLessonByLocations(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	}
	if userRole == "teacher" {
		token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.BobConn).RetrieveLiveLessonByLocations(helper.GRPCContext(ctx, "token", token), req)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) bobMustReturnLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkUserClassIDInLesson(ctx context.Context, lesson *pb.Lesson) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(lesson.UserClassIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	var count int64
	query := `SELECT COUNT(*) FROM courses_classes cc LEFT JOIN class_members cm ON cc.class_id = cm.class_id
		WHERE cc.course_id = $1 AND cm.user_id = $2 AND cm.class_id = ANY($3)`
	if err := s.BobDB.QueryRow(ctx, query, lesson.CourseId, &stepState.CurrentUserID, &lesson.UserClassIds).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("class_id return does not belong to lesson")
	}
	return StepStateToContext(ctx, stepState), nil
}

// validate gRPC response
func (s *Suite) bobReturnResultLiveLessonForStudentWithLocationID(ctx context.Context, result, userRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "empty" {
		return s.bpbBobMustReturnEmptyLiveLesson(ctx)
	}
	if result == "correct" {
		return s.bpbBobMustReturnCorrectLiveLessonsForUser(ctx, userRole)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bpbBobMustReturnEmptyLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	if len(rsp.Lessons) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return empty lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) bpbBobMustReturnCorrectLiveLessonsForUser(ctx context.Context, userRole string) (context.Context, error) {
	ctx, err1 := s.bpbBobMustReturnLessons(ctx)
	ctx, err2 := s.returnLessonsMustHaveCorrectStatus(ctx)
	ctx, err3 := s.returnLessonsMustHaveCorrectTeacherProfile(ctx)
	var err4 error
	if userRole == "student" {
		ctx, err4 = s.resultCorrectLessonMember(ctx)
	} else if userRole == "teacher" {
		// teacher
		ctx, err4 = s.resultCorrectCourseInLesson(ctx)
	}
	ctx, err5 := s.returnLessonsMustHaveCorrectSchedulingStatus(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5)
	return ctx, err
}
func (s *Suite) bpbBobMustReturnLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	if len(rsp.Lessons) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) returnLessonsMustHaveCorrectStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	for _, lesson := range rsp.Lessons {
		status := lesson.Status
		if lesson.EndTime.Nanos < int32(time.Now().Nanosecond()) && status == cpb.LessonStatus_LESSON_STATUS_COMPLETED {
			continue
		}
		if lesson.StartTime.Nanos >= int32(time.Now().Nanosecond()) && status == cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS {
			continue
		}
		if lesson.StartTime.Nanos < int32(time.Now().Nanosecond()) && status == cpb.LessonStatus_LESSON_STATUS_NOT_STARTED {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s return wrong status: %s", lesson.LessonId, status.String())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnLessonsMustHaveCorrectSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	lessonRepo := repo.LessonRepo{}
	for _, lesson := range rsp.Lessons {
		lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, lesson.LessonId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("get lesson by lesson Id %s fail: %w", lesson.LessonID, err)
		}
		if lesson.SchedulingStatus == domain.LessonSchedulingStatusPublished {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s return wrong status: %s", lesson.LessonID, lesson.SchedulingStatus)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) resultCorrectCourseInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.RetrieveLiveLessonByLocationsRequest)
	if len(req.CourseIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	for _, lesson := range rsp.GetLessons() {
		if !contains(req.CourseIds, lesson.CourseId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match course retrieve\n%s\n%s", lesson.LessonId, req.CourseIds, lesson.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func contains(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}
func (s *Suite) resultCorrectLessonMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	lessonIDs := []string{}
	for _, lesson := range rsp.Lessons {
		lessonIDs = append(lessonIDs, lesson.LessonId)
	}
	t, _ := jwt.ParseString(stepState.AuthToken)
	query := `SELECT result_lesson
	FROM UNNEST($1::TEXT[]) AS result_lesson
	LEFT JOIN lesson_members lm ON result_lesson=lm.lesson_id AND lm.user_id =$2
	WHERE lm.lesson_id IS NULL`
	rows, err := s.BobDB.Query(ctx, query, database.TextArray(lessonIDs), t.Subject())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	wrongLessons := []string{}
	for rows.Next() {
		var wrongLesson string
		if err := rows.Scan(&wrongLesson); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan :%v", err)
		}
		wrongLessons = append(wrongLessons, wrongLesson)
	}
	if len(wrongLessons) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student not a member of these lesson %s", wrongLessons)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) returnLessonsMustHaveCorrectTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLiveLessonByLocationsResponse)
	for _, lesson := range rsp.Lessons {
		if len(lesson.Teacher) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson return does not contain teacher")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
