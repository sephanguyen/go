package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ys_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

/* #nosec */
func (s *suite) schoolAdminDeleteALos(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.LoIDs == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("empty learning objective")
	}
	numOfDeletedLos := rand.Intn(len(stepState.LoIDs)-1) + 1
	stepState.DeletedLoIDs = nil
	for i := 0; i < numOfDeletedLos; i++ {
		stepState.DeletedLoIDs = append(stepState.DeletedLoIDs, stepState.LoIDs[i])
	}

	stepState.AuthToken = stepState.SchoolAdminToken
	_, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).DeleteLos(s.signedCtx(ctx), &epb.DeleteLosRequest{
		LoIds: stepState.DeletedLoIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the school admin unable to delete los: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminhasAddedTheNewCourseForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err := s.BobDB.QueryRow(ctx, stmt, stepState.StudentID).Scan(&studentEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	_, err = ypb.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(appendContextWithToken(ctx, stepState.SchoolAdminToken), &ypb.UpdateStudentRequest{
		SchoolId: stepState.SchoolIDInt,
		StudentProfile: &ypb.UpdateStudentRequest_StudentProfile{
			Id:               stepState.StudentID,
			Name:             fmt.Sprintf("student-name+%s", stepState.StudentID),
			Grade:            stepState.Grade,
			EnrollmentStatus: cpb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			Email:            studentEmail,
		},
		StudentPackageProfiles: []*ypb.UpdateStudentRequest_StudentPackageProfile{
			{
				Id: &ypb.UpdateStudentRequest_StudentPackageProfile_CourseId{
					CourseId: stepState.CourseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			},
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to add course to student: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasCreatedANewCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken

	stepState.CourseID = idutil.ULIDNow()
	_, err := ys_pb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(appendContextWithToken(ctx, stepState.SchoolAdminToken), &ys_pb.UpsertCoursesRequest{
		Courses: []*ys_pb.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_BIOLOGY,
				DisplayOrder: 1,
				SchoolId:     stepState.SchoolIDInt,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
			},
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.BookID = idutil.ULIDNow()
	stepState.BookIDs = append(stepState.BookIDs, stepState.BookID)
	if stepState.SchoolIDInt == 0 {
		stepState.SchoolIDInt = constants.ManabieSchool
	}
	bookReq := &epb.UpsertBooksRequest_Book{
		BookId: stepState.BookID,
		Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
	}

	if _, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(contextWithToken(s, ctx), &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			bookReq,
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedABook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = appendContextWithToken(ctx, stepState.SchoolAdminToken)

	if ctx, err := s.createBook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %w", err)
	}

	if ctx, err := s.schoolAdminCreateContentBook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("schoolAdminCreateContentBook: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) queryBooks(ctx context.Context, bookIDs []string) (_ context.Context, books []*entities.Book, _ error) {
	stepState := StepStateFromContext(ctx)

	bEnt := &entities.Book{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE book_id = ANY($1)",
		strings.Join(database.GetFieldNames(bEnt), ","), bEnt.TableName())

	rows, err := s.DB.Query(ctx, query, bookIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		book := &entities.Book{}
		if err := rows.Scan(database.GetScanFields(book, database.GetFieldNames(book))...); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		books = append(books, book)
	}

	return StepStateToContext(ctx, stepState), books, nil
}

func (s *suite) queryChapters(ctx context.Context, chapterIDs []string) (_ context.Context, chapters []*entities.Chapter, _ error) {
	stepState := StepStateFromContext(ctx)

	chapterEnt := &entities.Chapter{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chapter_id = ANY($1)",
		strings.Join(database.GetFieldNames(chapterEnt), ","), chapterEnt.TableName())

	rows, err := s.DB.Query(ctx, query, chapterIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		chapter := &entities.Chapter{}
		if err := rows.Scan(database.GetScanFields(chapter, database.GetFieldNames(chapter))...); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		chapters = append(chapters, chapter)
	}

	return StepStateToContext(ctx, stepState), chapters, nil
}

func (s *suite) queryBookChaptersByBookIDs(ctx context.Context, bookIDs []string) (_ context.Context, bookChapters []*entities.BookChapter, _ error) {
	stepState := StepStateFromContext(ctx)

	bookChapterEnt := &entities.BookChapter{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE book_id = ANY($1)",
		strings.Join(database.GetFieldNames(bookChapterEnt), ","), bookChapterEnt.TableName())

	rows, err := s.DB.Query(ctx, query, bookIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		bookChapter := &entities.BookChapter{}
		if err := rows.Scan(database.GetScanFields(bookChapter, database.GetFieldNames(bookChapter))...); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		bookChapters = append(bookChapters, bookChapter)
	}

	return StepStateToContext(ctx, stepState), bookChapters, nil
}

func (s *suite) queryTopicsByChapterIDs(ctx context.Context, chapterIDs []string) (_ context.Context, topics []*entities.Topic, _ error) {
	stepState := StepStateFromContext(ctx)

	topicEnt := &entities.Topic{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE chapter_id = ANY($1)",
		strings.Join(database.GetFieldNames(topicEnt), ","), topicEnt.TableName())

	rows, err := s.DB.Query(ctx, query, chapterIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		topic := &entities.Topic{}
		if err := rows.Scan(database.GetScanFields(topic, database.GetFieldNames(topic))...); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		topics = append(topics, topic)
	}

	return StepStateToContext(ctx, stepState), topics, nil
}

func (s *suite) queryLOsByTopicIDs(ctx context.Context, topicIDs []string) (_ context.Context, los []*entities.LearningObjective, _ error) {
	stepState := StepStateFromContext(ctx)

	loEnt := &entities.LearningObjective{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = ANY($1)",
		strings.Join(database.GetFieldNames(loEnt), ","), loEnt.TableName())

	rows, err := s.DB.Query(ctx, query, topicIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		lo := &entities.LearningObjective{}
		if err := rows.Scan(database.GetScanFields(lo, database.GetFieldNames(lo))...); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		los = append(los, lo)
	}

	return StepStateToContext(ctx, stepState), los, nil
}

func (s *suite) getStudentStudyPlanID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var studentStudyPlan string

	stmt := `SELECT study_plan_id FROM study_plans WHERE master_study_plan_id=$1 AND course_id=$2 ORDER BY created_at DESC LIMIT 1`
	err := s.DB.QueryRow(ctx, stmt, database.Text(stepState.StudyPlanID), database.Text(stepState.CourseID)).Scan(&studentStudyPlan)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentStudyPlanID = studentStudyPlan

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRetrieveStudentProgress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	ssp, err := (&repositories.StudentStudyPlanRepo{}).ListStudyPlans(ctx, s.DB, &repositories.ListStudyPlansArgs{
		StudentID: database.Text(stepState.StudentID),
		CourseID:  database.Text(stepState.CourseID),
		SchoolID:  pgtype.Int4{Status: pgtype.Null},
		Limit:     1,
		Offset:    pgtype.Text{Status: pgtype.Null},
	})

	if err != nil || len(ssp) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get student study plan id: %w", err)
	}

	stepState.AuthToken = stepState.TeacherToken
	stepState.Response, err = epb.NewAssignmentReaderServiceClient(s.Conn).RetrieveStudyPlanProgress(s.signedCtx(ctx), &epb.RetrieveStudyPlanProgressRequest{
		StudyPlanId: ssp[0].ID.String,
		StudentId:   stepState.StudentID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("teacher unable to retrieve study plan progress: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToReturnStudentProgressCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.RetrieveStudyPlanProgressResponse)
	if int(resp.GetTotalAssignments()) != len(stepState.AvailableStudyPlanIDs)-len(stepState.DeletedLoIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of total assignment: want %d, actual %d", len(stepState.AvailableStudyPlanIDs)-len(stepState.DeletedLoIDs), int(resp.GetTotalAssignments()))
	}
	return StepStateToContext(ctx, stepState), nil
}

// PREPARING DATA
func (s *suite) schoolAdminCreateAtopicAndAChapter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, pbChapters := s.prepareChapterV1(ctx, 1)
	if ctx, err := s.createChapterV1(ctx, stepState.BookID, pbChapters); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, pbTopics := s.prepareTopicInfoV1(ctx, stepState.ChapterID, 1)
	if ctx, err := s.createTopicV1(ctx, pbTopics); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminCreateContentBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, pbLOs := s.prepareLO(ctx, stepState.TopicID, 1, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING)
	if ctx, err := s.createLO(ctx, pbLOs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, pbLOs = s.prepareLO(ctx, stepState.TopicID, 1, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_FLASH_CARD)
	if ctx, err := s.createLO(ctx, pbLOs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, pbLOs = s.prepareLO(ctx, stepState.TopicID, 1, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_OFFLINE_LEARNING)
	if ctx, err := s.createLO(ctx, pbLOs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareTopicInfoV1(ctx context.Context, chapterID string, numberOfTopic int) (context.Context, []*epb.Topic) {
	stepState := StepStateFromContext(ctx)
	pbTopics := make([]*epb.Topic, 0, numberOfTopic)
	for i := 0; i < numberOfTopic; i++ {
		now := time.Now()
		stepState.TopicID = idutil.ULIDNow()
		stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
		pbTopics = append(pbTopics, &epb.Topic{
			Id:           stepState.TopicID,
			Name:         fmt.Sprintf("topic-name+%s", stepState.TopicID),
			Country:      epb.Country_COUNTRY_VN,
			Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
			Subject:      epb.Subject_SUBJECT_BIOLOGY,
			Type:         epb.TopicType_TOPIC_TYPE_ASSIGNMENT,
			CreatedAt:    timestamppb.New(now),
			UpdatedAt:    timestamppb.New(now),
			DisplayOrder: int32(i + 1),
			TotalLos:     1,
			ChapterId:    chapterID,
			SchoolId:     stepState.SchoolIDInt,
		})
	}
	return StepStateToContext(ctx, stepState), pbTopics
}

func (s *suite) prepareTopicInfo(ctx context.Context, chapterID string, numberOfTopic int) (context.Context, []*bob_pb.Topic) {
	stepState := StepStateFromContext(ctx)
	pbTopics := make([]*bob_pb.Topic, 0, numberOfTopic)
	for i := 0; i < numberOfTopic; i++ {
		now := time.Now()
		stepState.TopicID = idutil.ULIDNow()
		stepState.TopicIDs = append(stepState.TopicIDs, stepState.TopicID)
		pbTopics = append(pbTopics, &bob_pb.Topic{
			Id:           stepState.TopicID,
			Name:         fmt.Sprintf("topic-name+%s", stepState.TopicID),
			Country:      bob_pb.COUNTRY_VN,
			Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
			Subject:      bob_pb.SUBJECT_BIOLOGY,
			Type:         bob_pb.TOPIC_TYPE_LEARNING,
			CreatedAt:    &types.Timestamp{Seconds: now.Unix()},
			UpdatedAt:    &types.Timestamp{Seconds: now.Unix()},
			DisplayOrder: int32(i + 1),
			TotalLos:     1,
			ChapterId:    chapterID,
			SchoolId:     stepState.SchoolIDInt,
		})
	}
	return StepStateToContext(ctx, stepState), pbTopics
}

func (s *suite) createTopic(ctx context.Context, epbTopic []*epb.Topic) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: epbTopic,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topic: %w", err)
	}
	if resp.GetTopicIds() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create any topic")
	}
	// publish topic to event
	_, err = epb.NewTopicModifierServiceClient(s.Conn).Publish(ctx, &epb.PublishTopicsRequest{
		TopicIds: resp.GetTopicIds(),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to publish topic: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createTopicV1(ctx context.Context, pbTopic []*epb.Topic) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: pbTopic,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topic: %w", err)
	}
	if resp.GetTopicIds() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create any topic")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) prepareChapter(ctx context.Context, numberOfChapter int) (context.Context, []*ys_pb.Chapter) {
	stepState := StepStateFromContext(ctx)
	pbChapters := make([]*ys_pb.Chapter, 0, numberOfChapter)
	for i := 0; i < numberOfChapter; i++ {
		stepState.ChapterID = idutil.ULIDNow()
		stepState.ChapterIDs = append(stepState.ChapterIDs, stepState.ChapterID)
		pbChapters = append(pbChapters, &ys_pb.Chapter{
			ChapterId:    stepState.ChapterID,
			ChapterName:  fmt.Sprintf("chapter-name+%s", stepState.ChapterID),
			Country:      bob_pb.COUNTRY_VN,
			Subject:      bob_pb.SUBJECT_BIOLOGY,
			Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
			DisplayOrder: int32(i + 1),
			SchoolId:     stepState.SchoolIDInt,
		})
	}
	return StepStateToContext(ctx, stepState), pbChapters
}

func (s *suite) prepareChapterV1(ctx context.Context, numberOfChapter int) (context.Context, []*cpb.Chapter) {
	stepState := StepStateFromContext(ctx)
	pbChapters := make([]*cpb.Chapter, 0, numberOfChapter)
	for i := 0; i < numberOfChapter; i++ {
		stepState.ChapterID = idutil.ULIDNow()
		stepState.ChapterIDs = append(stepState.ChapterIDs, stepState.ChapterID)
		pbChapters = append(pbChapters, &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Id:           stepState.ChapterID,
				Name:         fmt.Sprintf("chapter-name+%s", stepState.ChapterID),
				Country:      cpb.Country_COUNTRY_VN,
				Subject:      cpb.Subject_SUBJECT_BIOLOGY,
				Grade:        stepState.Grade,
				DisplayOrder: int32(i + 1),
				SchoolId:     constants.ManabieSchool,
			},
		})
	}
	return StepStateToContext(ctx, stepState), pbChapters
}

func (s *suite) createChapter(ctx context.Context, bookID string, cpbChapters []*cpb.Chapter) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &epb.UpsertChaptersRequest{
		Chapters: cpbChapters,
		BookId:   bookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapter: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createChapterV1(ctx context.Context, bookID string, pbChapters []*cpb.Chapter) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: pbChapters,
		BookId:   bookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapter: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

// prepareLO create type learning.
// So remember we have more than one type on LO type, so create using random or which type use want with other func
func (s *suite) prepareLO(ctx context.Context, topicID string, numberOfLOs int, loType cpb.LearningObjectiveType) (context.Context, []*cpb.LearningObjective) {
	stepState := StepStateFromContext(ctx)
	pbLOs := make([]*cpb.LearningObjective, 0, numberOfLOs)
	for i := 0; i < numberOfLOs; i++ {
		stepState.LoID = idutil.ULIDNow()
		stepState.LoIDs = append(stepState.LoIDs, stepState.LoID)
		pbLOs = append(pbLOs, &cpb.LearningObjective{
			Info: &cpb.ContentBasicInfo{
				Id:           stepState.LoID,
				Name:         fmt.Sprintf("lo-%s-name+%s", loType.String(), stepState.LoID),
				Country:      cpb.Country_COUNTRY_VN,
				Grade:        1,
				Subject:      cpb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(i + 1),

				SchoolId: stepState.SchoolIDInt,
			},
			Type:          loType,
			TopicId:       topicID,
			Instruction:   "test instruction",
			GradeToPass:   wrapperspb.Int32(1),
			ManualGrading: true,
			TimeLimit:     wrapperspb.Int32(1),
		})
	}
	return StepStateToContext(ctx, stepState), pbLOs
}

func (s *suite) createLOEureka(ctx context.Context, pbLOs []*cpb.LearningObjective) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(ctx, &epb.UpsertLOsRequest{
		LearningObjectives: pbLOs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create learning objective: %w", err)
	}
	if resp.GetLoIds() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create LO: empty")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLO(ctx context.Context, pbLOs []*cpb.LearningObjective) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &epb.UpsertLOsRequest{
		LearningObjectives: pbLOs,
	}
	stepState.Request = req

	resp, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create learning objective: %w", err)
	}

	if resp.GetLoIds() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create LO: empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addsTheNewlyCreatedBookForCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.addsTheNewlyCreatedBookForCourseWithArgs(ctx, stepState.BookID, stepState.CourseID)
}

func (s *suite) addsTheNewlyCreatedBookForCourseWithArgs(ctx context.Context, bookID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = appendContextWithToken(ctx, stepState.SchoolAdminToken)
	_, err := epb.NewCourseModifierServiceClient(s.Conn).AddBooks(s.signedCtx(ctx), &epb.AddBooksRequest{
		BookIds:  []string{bookID},
		CourseId: courseID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to add a book to course: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) topicLearningObjectivesOfDeletedLosWereSuccessfullyDeleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.TopicsLearningObjectives{}
	deletedTopicLOs := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE lo_id = ANY($1::TEXT[]) AND deleted_at IS NOT NULL", e.TableName())

	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx, deletedTopicLOs, database.TextArray(stepState.DeletedLoIDs)).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable query count: %w", err)
	}

	if int(count.Int) != len(stepState.DeletedLoIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("los were not deleted")
	}

	return StepStateToContext(ctx, stepState), nil
}
