package exam_lo

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ys_pb_v1 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func (s *Suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// nolint
	addedRandom := "-" + strconv.Itoa(rand.Int())
	listLocation := []struct {
		locationID       string
		name             string
		parentLocationID string
		archived         bool
		expected         bool
		accessPath       string
	}{ // satisfied
		{locationID: "1" + addedRandom, parentLocationID: "", archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1"})},
		{locationID: "2" + addedRandom, parentLocationID: "1" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2"})},
		{locationID: "3" + addedRandom, parentLocationID: "2" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2", "3"})},
		{locationID: "7" + addedRandom, parentLocationID: "", archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7"})},
		// unsatisfied
		{locationID: "4" + addedRandom, parentLocationID: "", archived: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4"})},
		{locationID: "5" + addedRandom, parentLocationID: "4" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "6" + addedRandom, parentLocationID: "5" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "8" + addedRandom, parentLocationID: "7" + addedRandom, archived: true, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7", "8"})},
	}

	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived, access_path) VALUES($1,$2,$3,$4,$5) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, l.locationID,
			l.name,
			sql.NullString{},
			l.archived, l.accessPath)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) upsertCourse(ctx context.Context) string {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.aListOfLocationsInDB(ctx)
	if err != nil {
		return err.Error()
	}

	stepState.CourseID = idutil.ULIDNow()
	req, ok := stepState.Request.(*mpb.UpsertCoursesRequest)

	if !ok {
		req = &mpb.UpsertCoursesRequest{
			Courses: []*mpb.UpsertCoursesRequest_Course{},
		}
	}
	course := &mpb.UpsertCoursesRequest_Course{
		Id:             stepState.CourseID,
		Name:           "course-name-grade-book-" + stepState.CourseID,
		SchoolId:       constants.ManabieSchool,
		Icon:           "link-icon",
		LocationIds:    stepState.LocationIDs[0:2],
		TeachingMethod: mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_NONE,
	}
	req.Courses = append(req.Courses, course)
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.Request.(*mpb.UpsertCoursesRequest))
	return stepState.CourseID
}

func (s *Suite) userCreateABook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			{
				BookId: idutil.ULIDNow(),
				Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
			},
		},
	})
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not create a book %v", stepState.ResponseErr)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Create mock student login
func (s *Suite) studentLogin(ctx context.Context, numStudent int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := 0; i < numStudent; i++ {
		studentID, StudentToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, "student")
		stepState.StudentIDs = append(stepState.StudentIDs, studentID)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.Students = append(
			stepState.Students,
			entity.Student{
				ID:    studentID,
				Token: StudentToken,
			},
		)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher login
func (s *Suite) teacherLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	_, token, err := s.AuthHelper.AUserSignedInAsRole(ctx, "teacher")
	if err != nil {
		return nil, fmt.Errorf("error teacher login failed: %w", err)
	}
	stepState.TeacherToken = token

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock school admin login
func (s *Suite) schoolAdminLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	_, token, err := s.AuthHelper.AUserSignedInAsRole(ctx, "school admin")
	if err != nil {
		return nil, fmt.Errorf("error school admin login failed: %w", err)
	}
	stepState.SchoolAdminToken = token
	stepState.Token = stepState.SchoolAdminToken

	return utils.StepStateToContext(ctx, stepState), nil
}

// This step create mock data on the course
// base on number of provided by args
func (s *Suite) hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2(ctx context.Context, role string, numExamLos, gradeToPassPoint, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	var err error
	stepState := utils.StepStateFromContext[StepState](ctx)
	// add school_admin token to context
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	// Create 1 book
	bookGenerated, err := utils.GenerateBooksV2(ctx, 1, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %v", err)
	}
	stepState.BookID = bookGenerated.BookIDs[0]

	// Create content book
	if ctx, err = s.hasCreateContentBookForStudentProgress(ctx, numExamLos, gradeToPassPoint, numAsses, numTopics, numChapters, numQuizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminCreateACourseWithABook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Create course
	// Assume Course has create successfully in yasuo service
	// and return courseID
	ctx, err := s.aListOfLocationsInDB(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("aListOfLocationsInDB: %w", err)
	}

	stepState.CourseID = idutil.ULIDNow()
	req, ok := stepState.Request.(*mpb.UpsertCoursesRequest)

	if !ok {
		req = &mpb.UpsertCoursesRequest{
			Courses: []*mpb.UpsertCoursesRequest_Course{},
		}
	}
	course := &mpb.UpsertCoursesRequest_Course{
		Id:             stepState.CourseID,
		Name:           "course-name-grade-book-" + stepState.CourseID,
		SchoolId:       constants.ManabieSchool,
		Icon:           "link-icon",
		LocationIds:    stepState.LocationIDs[0:2],
		TeachingMethod: mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_NONE,
	}
	req.Courses = append(req.Courses, course)
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.Request.(*mpb.UpsertCoursesRequest))

	// Add book to couse
	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) hasCreateContentBookForStudentProgress(ctx context.Context, numExamLos, gradeToPassPoint, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	// Create Book chapters
	// Use school admin token for auth
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	chaptersGenerated, err := utils.GenerateChaptersV2(ctx, stepState.BookID, numChapters, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	for _, chapterID := range chaptersGenerated.ChapterIDs {
		// Create topic
		topicsGenerated, err := utils.GenerateTopicsV2(ctx, chapterID, numTopics, nil, s.EurekaConn)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		for _, topicID := range topicsGenerated.TopicIDs {
			// Create learning objective
			losGenerated, err := utils.GenerateLearningObjectivesTypeExamLo(ctx, topicID, numExamLos, gradeToPassPoint, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_EXAM_LO, nil, s.EurekaConn)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
			stepState.LoIDs = losGenerated.LoIDs

			for _, loID := range losGenerated.LoIDs {
				// Create quizz
				// Use school admin token for auth
				stepState.Token = stepState.SchoolAdminToken
				ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

				if err := utils.GenerateQuizzes(ctx, loID, numQuizzes, nil, s.EurekaConn); err != nil {
					return utils.StepStateToContext(ctx, stepState), err
				}
			}
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsCorrectTopicStatistic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := stepState.Request.(*epb.RetrieveCourseStatisticRequestV2)
	statisticResp := stepState.Response.(*epb.RetrieveCourseStatisticResponseV2)

	classID := pgtype.TextArray{Status: pgtype.Null}
	if len(req.ClassId) != 0 {
		classID = database.TextArray(req.ClassId)
	}

	items, err := (&repositories.CourseStudyPlanRepo{}).ListCourseStatisticItemsV2(ctx, s.EurekaDB, &repositories.ListCourseStatisticItemsArgsV2{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     classID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error repo list topic statistic items %w", err)
	}

	// order of topic, study plan is correct
	topicIDs := []string{}
	studyPlanItemIDs := []string{}
	for _, item := range items {
		topicIDs = append(topicIDs, item.ContentStructure.TopicID)
		studyPlanItemIDs = append(studyPlanItemIDs, item.RootStudyPlanItemID)
	}
	topicIDs = golibs.GetUniqueElementStringArray(topicIDs)
	studyPlanItemIDs = golibs.GetUniqueElementStringArray(studyPlanItemIDs)
	stepState.TopicIDs = topicIDs

	if len(topicIDs) != len(statisticResp.GetTopicStatistic()) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect %v topics got %v, courseID: %s, masterStudyPlamitemID: %s", len(topicIDs), len(statisticResp.GetTopicStatistic()), stepState.CourseID, stepState.StudyPlanID)
	}

	sort.Strings(stepState.ArchivedStudyPlanItemIDs)

	spiIndex := 0
	for i, statItem := range statisticResp.GetTopicStatistic() {
		if statItem.TopicId != topicIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong topic order")
		}

		for _, studyPlanItem := range statItem.GetLearningMaterialStatistic() {
			if spiIndex >= len(studyPlanItemIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("more study plan items return than expected")
			}
			if sort.SearchStrings(stepState.ArchivedStudyPlanItemIDs, studyPlanItem.GetStudyPlanItemId()) != len(stepState.ArchivedStudyPlanItemIDs) {
				// archived item
				if studyPlanItem.GetTotalAssignedStudent() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("total assigned student mustn't count archived item")
				}
				if studyPlanItem.GetCompletedStudent() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("completed student mustn't count archived item")
				}
				if studyPlanItem.GetAverageScore() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("average score mustn't include archived item")
				}
			}
			spiIndex++
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateCouseDurationForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	for _, student := range stepState.Students {
		query := `SELECT email FROM users WHERE user_id = $1`
		var studentEmail string
		err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
			ctx,
			&ys_pb_v1.UpdateStudentRequest{
				StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
					Id:               student.ID,
					Name:             "test-name",
					Grade:            5,
					EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            studentEmail,
				},
				SchoolId: int32(stepState.SchoolIDInt),
			},
		)
		_, err = ys_pb_v1.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
			StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
				Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
					CourseId: stepState.CourseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			}},
			StudentId: student.ID,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course duration: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// Create mock studyplan data
func (s *Suite) hasCreatedAStudyplanForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	// This studyPlanID is also a masterStudyPlanID
	generatedStudyPlan, err := utils.GenerateStudyPlanV2(ctx, s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = generatedStudyPlan.StudyPlanID

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range stepState.StudyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		err = item.ContentStructure.AssignTo(cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		if len(cse.LoID) != 0 {
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}
	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, numStudent int, arg1 int, arg2 int, arg3 int, arg4 int, arg5 int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.Students) < numStudent {
		stepState.Students = stepState.Students[0:numStudent]
	}

	stepState.ShuffledQuizSetIDs = []string{}
	for i := 0; i < numStudent; i++ {
		stepState.StudentID = stepState.Students[i].ID
		stepState.StudentToken = stepState.Students[i].Token
		ctx, shuffledQuizSetIDs, err := s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(utils.StepStateToContext(ctx, stepState), arg1, arg2, arg3, arg4, arg5)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.ShuffledQuizSetIDs = append(stepState.ShuffledQuizSetIDs, shuffledQuizSetIDs...)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, numWorkDoneLO int, numCorrectQuiz int, numWorkDoneAss int, assPoint int, skippedTopic int) (context.Context, []string, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	shuffledQuizSetIDs := []string{}

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.EurekaDB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to fetch study plan items: %w", err)
	}

	chapters := make(map[string][]string)

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		if err := each.ContentStructure.AssignTo(cs); err != nil {
			return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("error assignto ContentStructure for chapter id")
		}
		chapters[cs.ChapterID] = append(chapters[cs.ChapterID], cs.TopicID)
	}

	for _, each := range chapters {
		for i := 0; i < skippedTopic; i++ {
			stepState.SkippedTopics = append(stepState.SkippedTopics, each[i])
		}
	}

	for i := 0; i < numWorkDoneLO; i++ {
		loID := stepState.LoIDs[i]
		for _, each := range studyPlanItems {
			cs := new(entities.ContentStructure)
			if err := each.ContentStructure.AssignTo(cs); err != nil {
				return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("error assignto ContentStructure in studyplan")
			}
			if utils.ContainsStr(stepState.SkippedTopics, cs.TopicID) {
				continue
			}

			if cs.LoID == loID {
				// Use student token for auth
				stepState.Token = stepState.StudentToken
				ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
				resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(ctx, &epb.CreateQuizTestRequest{
					LoId:            cs.LoID,
					StudentId:       stepState.StudentID,
					StudyPlanItemId: each.ID.String,
					KeepOrder:       false,
					Paging: &cpb.Paging{
						Limit: uint32(100),
						Offset: &cpb.Paging_OffsetInteger{
							OffsetInteger: int64(1),
						},
					},
				})
				if err != nil {
					return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to create quiz test: %w", err)
				}
				stepState.ShuffledQuizSetID = resp.QuizzesId
				shuffledQuizSetIDs = append(shuffledQuizSetIDs, resp.QuizzesId)
				lNumCorrectQuiz := numCorrectQuiz
				for _, quiz := range resp.Items {
					if lNumCorrectQuiz > 0 {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "A"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to check quiz correctness: %w", err)
						}
					} else {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "B"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable to check quiz correctness: %w", err)
						}
						stepState.WrongQuizExternalIDs = append(stepState.WrongQuizExternalIDs, quiz.Core.ExternalId)
					}
					lNumCorrectQuiz--
				}

				query := `UPDATE study_plan_items set completed_at = now() WHERE study_plan_item_id = $1::TEXT`
				if _, err := s.EurekaDB.Exec(ctx, query, each.ID.String); err != nil {
					return utils.StepStateToContext(ctx, stepState), nil, fmt.Errorf("unable update complete study plan items: %v", err)
				}
			}

		}
	}
	return utils.StepStateToContext(ctx, stepState), shuffledQuizSetIDs, nil
}

func (s *Suite) waitForImportStudentStudyPlanCompleted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}

	err := try.Do(func(attempt int) (bool, error) {
		findByStudentIDsResp, err := studentStudyPlanRepo.FindByStudentIDs(ctx, s.EurekaDB, database.TextArray([]string{stepState.StudentID}))
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(time.Second)
				return attempt < 10, fmt.Errorf("no row found")
			}
			return false, err
		}

		if len(findByStudentIDsResp) != 0 {
			return false, nil
		}

		time.Sleep(time.Second)
		return attempt < 10, fmt.Errorf("timeout sync import student study plan")
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("waitForImportStudentStudyPlanCompleted error: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher retrieve topic statistical
func (s *Suite) retrieveTopicStatistic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use teacher token for auth
	stepState.Token = stepState.TeacherToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	req := &epb.RetrieveCourseStatisticRequestV2{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
		ClassId:     []string{},
	}

	if len(stepState.ClassIDs) != 0 {
		req.ClassId = stepState.ClassIDs
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewCourseReaderServiceClient(s.EurekaConn).RetrieveCourseStatisticV2(ctx, req)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error RetrieveCourseStatisticV2: %w", stepState.ResponseErr)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs(ctx context.Context, topicAssigned, topicCompleted, topicAverageScore int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	statisticResp := stepState.Response.(*epb.RetrieveCourseStatisticResponseV2)

	if len(statisticResp.TopicStatistic) == 0 {
		if topicAssigned != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = 0 got = %v", topicAssigned)
		}
		if topicCompleted != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = 0 got = %v", topicCompleted)
		}

		if topicAverageScore != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
		}
		return utils.StepStateToContext(ctx, stepState), nil
	}

	if topicAssigned != statisticResp.TopicStatistic[0].TotalAssignedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = %v got = %v", topicAssigned, statisticResp.TopicStatistic[0].TotalAssignedStudent)
	}

	if topicCompleted != statisticResp.TopicStatistic[0].CompletedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = %v got = %v", topicCompleted, statisticResp.TopicStatistic[0].CompletedStudent)
	}

	if topicAverageScore != statisticResp.TopicStatistic[0].AverageScore {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AssignClassToCourse(ctx context.Context, createFlag string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if createFlag == "true" {
		cs := []*entities.CourseClass{{
			ID:       database.Text(stepState.CourseID + stepState.ClassIDs[0]),
			CourseID: database.Text(stepState.CourseID),
			ClassID:  database.Text(stepState.ClassIDs[0])}}
		repo := repositories.CourseClassRepo{}
		err := repo.BulkUpsert(ctx, s.EurekaDB, cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) retrieveGradeBookWithSetting(ctx context.Context, setting string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	time.Sleep(2 * time.Second)
	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).ListGradeBook(s.AuthHelper.SignedCtx(metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0"), stepState.SchoolAdminToken), &sspb.GradeBookRequest{
		StudentIds:   stepState.StudentIDs,
		Setting:      sspb.GradeBookSetting(sspb.GradeBookSetting_value[setting]),
		CourseIds:    []string{},
		StudyPlanIds: []string{},
		Grades:       []int32{},
		LocationIds:  stepState.LocationIDs[0:2],
	})
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not retrieve grade book with setting %v", stepState.ResponseErr.Error())
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) teachUpdateResultOfExamLo(ctx context.Context, result string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := fmt.Sprintf(`
	update exam_lo_submission set result = '%s' where shuffled_quiz_set_id = any($1::TEXT[])
	`, result)

	if _, err := s.EurekaDB.Exec(ctx, query, &stepState.ShuffledQuizSetIDs); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not update result of exam lo: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnsGradeBookViewCorrectly(ctx context.Context, numStudentGrade, examLos, completedLos, numGradeToPass, passed, numExamResults, point, gradePoint, totalAttempts int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*sspb.GradeBookResponse)
	if len(resp.StudentGradeItems) != numStudentGrade {
		fmt.Printf("----------- %v %v\n", len(resp.StudentGradeItems), numStudentGrade)
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not match student grade items")
	}
	for _, item := range resp.StudentGradeItems {
		if len(item.GetResults()) != numExamResults {
			fmt.Printf("----------- %v %v\n", len(item.GetResults()), numExamResults)
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not match exam result items")
		}
		if int(item.GetTotalExamLos().GetValue()) != examLos ||
			int(item.GetTotalCompletedExamLos().GetValue()) != completedLos ||
			int(item.GetTotalGradeToPass().GetValue()) != numGradeToPass ||
			int(item.GetTotalPassed().GetValue()) != passed {
			fmt.Printf("-------- item ----- examLo %v completedExamLos %v TotalgradeToPass %v TotalPassed %v \n", int(item.GetTotalExamLos().GetValue()), int(item.GetTotalCompletedExamLos().GetValue()), int(item.GetTotalGradeToPass().GetValue()), int(item.GetTotalPassed().GetValue()))
			fmt.Printf("----- examLo %v completedExamLos %v TotalgradeToPass %v TotalPassed %v \n", examLos, completedLos, numGradeToPass, passed)
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not match exam result total score")
		}

		for _, examResult := range item.GetResults() {
			if int(examResult.GetTotalPoint().GetValue()) != point ||
				int(examResult.GetGradePoint().GetValue()) != gradePoint ||
				int(examResult.GetTotalAttempts()) != totalAttempts {
				fmt.Printf("-------- item ----- point %v gradePoint %v totalAttempts %v isFailed %v \n", int(examResult.GetTotalPoint().GetValue()), int(examResult.GetGradePoint().GetValue()), int(examResult.GetTotalAttempts()), examResult.Failed)
				fmt.Printf("------------- point %v gradePoint %v totalAttempts %v  \n", point, gradePoint, totalAttempts)
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not match exam result score")
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpsertGradeBookSetting(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).UpsertGradeBookSetting(s.AuthHelper.SignedCtx(metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0"), stepState.Token), &sspb.UpsertGradeBookSettingRequest{
		Setting: sspb.GradeBookSetting_GRADE_TO_PASS_SCORE,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminRespectivelyCreate3BooksWith0ExamLO1ExamLOAnd2ExamLOs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if ctx, err := s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2(ctx, "admin", 0, 10, 0, 0, 0, 0); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}
	stepState.BookIDs = append(stepState.BookIDs, stepState.BookID)
	for i := 1; i < 3; i++ {
		if ctx, err := s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2(ctx, "admin", i, 10, 1, 1, 1, 1); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
		}
		stepState.BookIDs = append(stepState.BookIDs, stepState.BookID)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreateCourseWith3StudyPlanUsing3Books(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	stepState.CourseID = s.upsertCourse(ctx)
	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  stepState.BookIDs,
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}

	for _, bookID := range stepState.BookIDs {
		generatedStudyPlan, err := utils.GenerateStudyPlanV2(ctx, s.EurekaConn, stepState.CourseID, bookID)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.StudyPlanID = generatedStudyPlan.StudyPlanID
	}

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range stepState.StudyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		err = item.ContentStructure.AssignTo(cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		if len(cse.LoID) != 0 {
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}
	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate2ndCourseWithStudyPlanHave2ExamLOs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.CourseID = s.upsertCourse(ctx)
	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookIDs[2]},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}

	generatedStudyPlan, err := utils.GenerateStudyPlanV2(ctx, s.EurekaConn, stepState.CourseID, stepState.BookIDs[2])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = generatedStudyPlan.StudyPlanID

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range stepState.StudyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		err = item.ContentStructure.AssignTo(cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		if len(cse.LoID) != 0 {
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}
	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate3rdCourseWithStudyPlanHave0ExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.CourseID = s.upsertCourse(ctx)
	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookIDs[0]},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}

	generatedStudyPlan, err := utils.GenerateStudyPlanV2(ctx, s.EurekaConn, stepState.CourseID, stepState.BookIDs[0])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = generatedStudyPlan.StudyPlanID

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range stepState.StudyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		err = item.ContentStructure.AssignTo(cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		if len(cse.LoID) != 0 {
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}
	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate4thCourseWithNoStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.CourseID = s.upsertCourse(ctx)
	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate1stStudentAtGrade5JoinAllCourses(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	student := stepState.Students[0]
	query := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
		ctx,
		&ys_pb_v1.UpdateStudentRequest{
			StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
				Id:               student.ID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
			},
			SchoolId: int32(stepState.SchoolIDInt),
		},
	)
	for _, courseID := range stepState.CourseIDs {
		_, err = ys_pb_v1.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
			StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
				Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
					CourseId: courseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			}},
			StudentId: student.ID,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course duration: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate2ndStudentAtGrade6Join1stCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	student := stepState.Students[1]
	query := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
		ctx,
		&ys_pb_v1.UpdateStudentRequest{
			StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
				Id:               student.ID,
				Name:             "test-name",
				Grade:            6,
				EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
			},
			SchoolId: int32(stepState.SchoolIDInt),
		},
	)
	_, err = ys_pb_v1.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
		StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseIDs[0],
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
		StudentId: student.ID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course duration: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate3rdStudentAtGrade7Join3rdCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	student := stepState.Students[2]
	query := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
		ctx,
		&ys_pb_v1.UpdateStudentRequest{
			StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
				Id:               student.ID,
				Name:             "test-name",
				Grade:            7,
				EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
			},
			SchoolId: int32(stepState.SchoolIDInt),
		},
	)
	_, err = ys_pb_v1.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
		StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseIDs[2],
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
		StudentId: student.ID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course duration: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminCreate4thStudentAtGrade5(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
	student := stepState.Students[3]
	query := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
		ctx,
		&ys_pb_v1.UpdateStudentRequest{
			StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
				Id:               student.ID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
			},
			SchoolId: int32(stepState.SchoolIDInt),
		},
	)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminGetListGradeBookWith(ctx context.Context, course string, grade string, student string, record_per_page string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	time.Sleep(2 * time.Second)
	var gradeInt32 []int32
	if len(strings.TrimSpace(grade)) != 0 {
		for _, v := range strings.Split(grade, ",") {
			g, err := strconv.Atoi(v)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot get grade parameter: %w", err)
			}
			gradeInt32 = append(gradeInt32, int32(g))
		}
	} else {
		gradeInt32 = nil
	}

	req := &sspb.GradeBookRequest{
		StudentIds: getElements(stepState.StudentIDs, student),
		CourseIds:  getElements(stepState.CourseIDs, course),
	}
	if len(strings.TrimSpace(record_per_page)) != 0 {
		limit, _ := strconv.Atoi(record_per_page)
		req.Paging = &cpb.Paging{
			Limit: uint32(limit),
		}
	}
	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).ListGradeBook(s.AuthHelper.SignedCtx(metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0"), stepState.SchoolAdminToken), req)
	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not retrieve grade book with setting %v", stepState.ResponseErr.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func getElements(arr []string, s string) []string {
	if len(s) == 0 {
		return arr
	}
	var result []string
	if len(s) == 0 {
		return nil
	}
	indexes := strings.Split(s, ",")
	for _, v := range indexes {
		ind, err := strconv.Atoi(v)
		if err != nil {
			return nil
		}
		result = append(result, arr[ind-1])
	}
	return result
}

func (s *Suite) ReturnsCorrectResponseStudentAndTotalItems(ctx context.Context, totalItems int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	gradeBookResponse := stepState.Response.(*sspb.GradeBookResponse)
	if len(gradeBookResponse.StudentGradeItems) != totalItems {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("number of records returned is incorrect: expected: %d, actual: %d", totalItems, len(gradeBookResponse.StudentGradeItems))
	}
	student := make(map[string]struct{})
	for _, item := range gradeBookResponse.StudentGradeItems {
		if _, ok := student[item.StudentId]; !ok {
			student[item.StudentId] = struct{}{}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
