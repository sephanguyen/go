package utils

import (
	"context"
	crypRand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuoPb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ys_pb_v1 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Block struct {
	Key               string      `json:"key"`
	Text              string      `json:"text"`
	Type              string      `json:"type"`
	Depth             int32       `json:"depth"`
	InlineStyleRanges []string    `json:"inlineStyleRanges"`
	EntityRanges      []string    `json:"entityRanges"`
	Data              interface{} `json:"data"`
}

type Data struct {
	Data string `json:"data"`
}

type Raw struct {
	Blocks    []Block     `json:"blocks"`
	EntityMap interface{} `json:"entityMap"`
}

func GenerateBooks(numberOfBooks int, template *epb.UpsertBooksRequest_Book) []*epb.UpsertBooksRequest_Book {
	if template == nil {
		// A valid create book req template
		template = &epb.UpsertBooksRequest_Book{
			Name: "Syllabus" + idutil.ULIDNow(),
		}
	}
	books := make([]*epb.UpsertBooksRequest_Book, 0)
	for i := 0; i < numberOfBooks; i++ {
		books = append(books, proto.Clone(template).(*epb.UpsertBooksRequest_Book))
	}
	return books
}

type GenerateBookResult struct {
	BookPb  []*epb.UpsertBooksRequest_Book
	BookIDs []string
}

func GenerateBooksV2(ctx context.Context, numberOfBooks int, template *epb.UpsertBooksRequest_Book, eurekaConn *grpc.ClientConn) (*GenerateBookResult, error) {
	books := make([]*epb.UpsertBooksRequest_Book, 0)
	for i := 0; i < numberOfBooks; i++ {
		book := template
		if template == nil {
			// A valid create book req template
			book = &epb.UpsertBooksRequest_Book{
				Name: "Syllabus" + idutil.ULIDNow(),
			}
		}
		books = append(books, book)
	}
	resp, err := epb.NewBookModifierServiceClient(eurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{Books: books})
	if err != nil {
		return nil, fmt.Errorf("school admin unable to create a book: %w\n %v", err, template)
	}

	result := &GenerateBookResult{BookPb: books, BookIDs: resp.GetBookIds()}
	return result, nil
}

func GenerateChapters(bookID string, numberOfChapter int, template *cpb.Chapter) []*cpb.Chapter {
	if template == nil {
		template = &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Country:  cpb.Country_COUNTRY_VN,
				SchoolId: constants.ManabieSchool,
				Subject:  cpb.Subject_SUBJECT_BIOLOGY,
				Grade:    1,
				Name:     "Chapter" + idutil.ULIDNow(),
			},
			BookId: bookID,
		}
	}
	chapters := make([]*cpb.Chapter, 0)
	for i := 0; i < numberOfChapter; i++ {
		chapters = append(chapters, proto.Clone(template).(*cpb.Chapter))
	}
	return chapters
}

type GenerateChaptersResult struct {
	ChapterPb  []*cpb.Chapter
	ChapterIDs []string
}

func GenerateChaptersV2(ctx context.Context, bookID string, numberOfChapter int, template *cpb.Chapter, eurekaConn *grpc.ClientConn) (*GenerateChaptersResult, error) {
	chapters := make([]*cpb.Chapter, 0)
	for i := 0; i < numberOfChapter; i++ {
		chapter := template
		if template == nil {
			chapter = &cpb.Chapter{
				Info: &cpb.ContentBasicInfo{
					Country:  cpb.Country_COUNTRY_VN,
					SchoolId: constants.ManabieSchool,
					Subject:  cpb.Subject_SUBJECT_BIOLOGY,
					Grade:    1,
					Name:     "Chapter" + idutil.ULIDNow(),
				},
				BookId: bookID,
			}
		}
		chapters = append(chapters, chapter)
	}
	resp, err := epb.NewChapterModifierServiceClient(eurekaConn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: chapters,
		BookId:   bookID,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create chapter: %w", err)
	}

	result := &GenerateChaptersResult{
		ChapterPb:  chapters,
		ChapterIDs: resp.GetChapterIds(),
	}

	return result, nil
}

func GenerateTopic(chapterID string, template *epb.Topic) *epb.Topic {
	if template == nil {
		template = &epb.Topic{
			SchoolId:  constants.ManabieSchool,
			Subject:   epb.Subject_SUBJECT_BIOLOGY,
			Name:      "Topic" + idutil.ULIDNow(),
			ChapterId: chapterID,
			Status:    epb.TopicStatus_TOPIC_STATUS_NONE,
			Type:      epb.TopicType_TOPIC_TYPE_LEARNING,
			TotalLos:  1,
		}
	}

	return proto.Clone(template).(*epb.Topic)
}

func GenerateTopics(chapterID string, numberOfTopics int, template *epb.Topic) []*epb.Topic {
	books := make([]*epb.Topic, 0)
	for i := 0; i < numberOfTopics; i++ {
		books = append(books, GenerateTopic(chapterID, template))
	}
	return books
}

type GenerateTopicResult struct {
	TopicPb  []*epb.Topic
	TopicIDs []string
}

func GenerateTopicsV2(ctx context.Context, chapterID string, numberOfTopics int, template *epb.Topic, eurekaConn *grpc.ClientConn) (*GenerateTopicResult, error) {
	topics := make([]*epb.Topic, 0)
	for i := 0; i < numberOfTopics; i++ {
		topic := template
		if template == nil {
			topic = &epb.Topic{
				SchoolId:     constants.ManabieSchool,
				Subject:      epb.Subject_SUBJECT_BIOLOGY,
				Name:         "Topic" + idutil.ULIDNow(),
				ChapterId:    chapterID,
				Status:       epb.TopicStatus_TOPIC_STATUS_NONE,
				Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
				TotalLos:     1,
				DisplayOrder: int32(i + 1),
			}
		}
		topics = append(topics, topic)
	}
	resp, err := epb.NewTopicModifierServiceClient(eurekaConn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: topics,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to generate topic: %w", err)
	}

	result := &GenerateTopicResult{TopicPb: topics, TopicIDs: resp.GetTopicIds()}

	return result, nil
}

// AValidBookContent ensure have at least 1 chapter, 1 topic
func AValidBookContent(ctx context.Context, eurekaConn *grpc.ClientConn, db database.Ext, schoolID int32) (bookID string, chapterIDs []string, topicIDs []string, err error) {
	bookResp, err := epb.NewBookModifierServiceClient(eurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: GenerateBooks(1, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewBookModifierService.UpsertBooks: %w", err)
		return
	}

	bookID = bookResp.BookIds[0]
	chapterResp, err := epb.NewChapterModifierServiceClient(eurekaConn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: GenerateChapters(bookID, 1, nil),
		BookId:   bookID,
	})
	if err != nil {
		err = fmt.Errorf("NewChapterModifierService.UpsertChapters: %w", err)

		return
	}
	chapterIDs = chapterResp.GetChapterIds()
	topicResp, err := epb.NewTopicModifierServiceClient(eurekaConn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: GenerateTopics(chapterIDs[0], 1, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewTopicModifierService.Upsert: %w", err)
		return
	}
	topicIDs = topicResp.GetTopicIds()
	return
}

func GenerateLearningObjective(topicID string) *cpb.LearningObjective {
	id := idutil.ULIDNow()

	return &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:        id,
			Name:      "learning",
			Country:   cpb.Country_COUNTRY_VN,
			Grade:     12,
			Subject:   cpb.Subject_SUBJECT_MATHS,
			MasterId:  "",
			SchoolId:  constants.ManabieSchool,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
		TopicId: topicID,
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide: "https://guides/1/master",
		Video:      "https://videos/1/master",
	}
}

type GenerateLearningObjectivesResult struct {
	LosPb []*cpb.LearningObjective
	LoIDs []string
}

func GenerateLearningObjectivesTypeExamLo(ctx context.Context, topicID string, numberOfLos, gradeToPassPoint int, loType cpb.LearningObjectiveType, template *cpb.LearningObjective, eurekaConn *grpc.ClientConn) (*GenerateLearningObjectivesResult, error) {
	los := make([]*cpb.LearningObjective, 0)
	for i := 0; i < numberOfLos; i++ {
		lo := template
		if template == nil {
			lo = &cpb.LearningObjective{
				Info: &cpb.ContentBasicInfo{
					Name:         "learning",
					Country:      cpb.Country_COUNTRY_VN,
					Grade:        12,
					Subject:      cpb.Subject_SUBJECT_MATHS,
					MasterId:     "",
					SchoolId:     constants.ManabieSchool,
					DisplayOrder: int32(i + 1),
					CreatedAt:    nil,
					UpdatedAt:    nil,
				},
				TopicId: topicID,
				Prerequisites: []string{
					"AL-PH3.1", "AL-PH3.2",
				},
				Type:       loType,
				StudyGuide: "https://guides/1/master",
				Video:      "https://videos/1/master",
				GradeToPass: &wrapperspb.Int32Value{
					Value: int32(gradeToPassPoint),
				},
				ReviewOption: cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY,
			}
		}
		los = append(los, lo)
	}
	resp, err := epb.NewLearningObjectiveModifierServiceClient(eurekaConn).UpsertLOs(ctx, &epb.UpsertLOsRequest{
		LearningObjectives: los,
	})
	if err != nil {
		return nil, fmt.Errorf("generate learning objective error: %w", err)
	}
	if resp.GetLoIds() == nil {
		return nil, fmt.Errorf("unable create LO: empty")
	}
	result := &GenerateLearningObjectivesResult{LosPb: los, LoIDs: resp.GetLoIds()}

	return result, nil
}

func GenerateLearningObjectivesV2(ctx context.Context, topicID string, numberOfLos int, loType cpb.LearningObjectiveType, template *cpb.LearningObjective, eurekaConn *grpc.ClientConn) (*GenerateLearningObjectivesResult, error) {
	los := make([]*cpb.LearningObjective, 0)
	for i := 0; i < numberOfLos; i++ {
		lo := template
		if template == nil {
			lo = &cpb.LearningObjective{
				Info: &cpb.ContentBasicInfo{
					Name:         "learning",
					Country:      cpb.Country_COUNTRY_VN,
					Grade:        12,
					Subject:      cpb.Subject_SUBJECT_MATHS,
					MasterId:     "",
					SchoolId:     constants.ManabieSchool,
					DisplayOrder: int32(i + 1),
					CreatedAt:    nil,
					UpdatedAt:    nil,
				},
				TopicId: topicID,
				Prerequisites: []string{
					"AL-PH3.1", "AL-PH3.2",
				},
				Type:       loType,
				StudyGuide: "https://guides/1/master",
				Video:      "https://videos/1/master",
			}
		}
		los = append(los, lo)
	}
	resp, err := epb.NewLearningObjectiveModifierServiceClient(eurekaConn).UpsertLOs(ctx, &epb.UpsertLOsRequest{
		LearningObjectives: los,
	})
	if err != nil {
		return nil, fmt.Errorf("generate learning objective error: %w", err)
	}
	if resp.GetLoIds() == nil {
		return nil, fmt.Errorf("unable create LO: empty")
	}
	result := &GenerateLearningObjectivesResult{LosPb: los, LoIDs: resp.GetLoIds()}

	return result, nil
}

func GenerateStudentEventLogs(ctx context.Context, log []*epb.StudentEventLog, eurekaConn *grpc.ClientConn) error {
	req := &epb.CreateStudentEventLogsRequest{
		StudentEventLogs: log,
	}
	resp, err := epb.NewStudentEventLogModifierServiceClient(eurekaConn).CreateStudentEventLogs(ctx, req)
	if err != nil {
		return fmt.Errorf("NewStudentEventLogModifierServiceClient.CreateStudentEventLogs: %w", err)
	}
	if !resp.Successful {
		return fmt.Errorf("error student_event_logs insert unsuccessful")
	}

	return nil
}

func GenerateFlashcard(ctx context.Context, eurekaConn *grpc.ClientConn, topicID string) (learningMaterialIDs []string, err error) {
	insertFlashcardReq := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: topicID,
				Name:    fmt.Sprintf("flashcard-name+%v", topicID),
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(eurekaConn).InsertFlashcard(ctx, insertFlashcardReq)
	if err != nil {
		err = fmt.Errorf("NewFlashCardClient.InsertFlashcard: %w", err)
		return
	}
	learningMaterialIDs = append(learningMaterialIDs, resp.LearningMaterialId)
	return learningMaterialIDs, nil
}

func GenerateCourse(ctx context.Context, yasuoConn *grpc.ClientConn) (courseID string, err error) {
	courseID = idutil.ULIDNow()
	if _, err = yasuoPb.NewCourseServiceClient(yasuoConn).UpsertCourses(ctx, &yasuoPb.UpsertCoursesRequest{
		Courses: []*yasuoPb.UpsertCoursesRequest_Course{
			{
				Id:       courseID,
				Name:     "course",
				Country:  1,
				Subject:  bpb.SUBJECT_BIOLOGY,
				SchoolId: constants.ManabieSchool,
			},
		},
	}); err != nil {
		err = fmt.Errorf("NewCourseServiceClient.UpsertCourses: %w", err)
		return courseID, err
	}
	return
}

type GenerateStudyPlanResult struct{ StudyPlanID string }

func GenerateStudyPlan(ctx context.Context, eurekaConn *grpc.ClientConn, courseID string, bookID string) (studyPlanID string, err error) {
	if resp, err := epb.NewCourseModifierServiceClient(eurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: courseID,
		BookIds:  []string{bookID},
	}); err != nil || !resp.Successful {
		return studyPlanID, fmt.Errorf("NewCourseModifierServiceClient.AddBooks: %w", err)
	}

	resp, err := epb.NewStudyPlanModifierServiceClient(eurekaConn).UpsertStudyPlan(ctx, &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", bookID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2, 3},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              bookID,
		CourseId:            courseID,
	})
	if err != nil {
		return "", fmt.Errorf("NewStudyPlanModifierServiceClient.UpsertStudyPlan: %w", err)
	}
	studyPlanID = resp.StudyPlanId
	return
}

func GenerateStudyPlanV2(ctx context.Context, eurekaConn *grpc.ClientConn, courseID string, bookID string) (*GenerateStudyPlanResult, error) {
	resp, err := epb.NewStudyPlanModifierServiceClient(eurekaConn).UpsertStudyPlan(ctx, &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", bookID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              bookID,
		CourseId:            courseID,
	})
	if err != nil {
		err = fmt.Errorf("NewStudyPlanModifierServiceClient.UpsertStudyPlan: %w", err)
		return nil, err
	}
	result := &GenerateStudyPlanResult{StudyPlanID: resp.StudyPlanId}

	return result, nil
}

func InsertUserIntoBob(ctx context.Context, bobDB database.Ext, userID string) error {
	now := time.Now()
	user := &bob_entities.User{}
	database.AllNullEntity(user)
	userName := "valid-user-import-by-eureka" + userID
	num := idutil.ULIDNow()

	err := multierr.Combine(
		user.Country.Set("COUNTRY_VN"),
		user.PhoneNumber.Set(fmt.Sprintf("+849%s", num)),
		user.Email.Set(fmt.Sprintf("valid-%s@email.com", num)),
		user.LastName.Set(userName),
		user.Group.Set("USER_GROUP_STUDENT"),
		user.ID.Set(userID),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.ResourcePath.Set(fmt.Sprintf("%d", constants.ManabieSchool)),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	orgID := constants.ManabieSchool
	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(orgID))
	_, err = database.Insert(ctx, user, bobDB.Exec)
	if err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func InsertMultiUserIntoBob(ctx context.Context, bobDB database.Ext, numOfUsers int) ([]string, error) {
	studentIDs := make([]string, 0, numOfUsers)
	for i := 0; i < numOfUsers; i++ {
		userID := idutil.ULIDNow()
		if err := InsertUserIntoBob(ctx, bobDB, userID); err != nil {
			return nil, fmt.Errorf("insertMultiUserIntoBob: %w", err)
		}
		studentIDs = append(studentIDs, userID)
	}
	return studentIDs, nil
}

func GenerateCourseStudyPlan(ctx context.Context, studyPlanID, courseID string, eurekaDB database.Ext) (*entities.CourseStudyPlan, error) {
	var c entities.CourseStudyPlan
	database.AllNullEntity(&c)
	now := timeutil.Now()
	err := multierr.Combine(
		c.StudyPlanID.Set(studyPlanID),
		c.CourseID.Set(courseID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	repo := repositories.CourseStudyPlanRepo{}
	csp := []*entities.CourseStudyPlan{&c}
	if err := repo.BulkUpsert(ctx, eurekaDB, csp); err != nil {
		return nil, err
	}
	return &c, nil
}

func generateCourseByStudentID(studentID, courseID string) (*entities.CourseStudent, error) {
	var c entities.CourseStudent
	database.AllNullEntity(&c)
	now := timeutil.Now()
	err := multierr.Combine(
		c.ID.Set(ksuid.New().String()),
		c.CourseID.Set(courseID),
		c.StudentID.Set(studentID),
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
		c.StartAt.Set(now.Add(-time.Hour)),
		c.EndAt.Set(now.Add(time.Hour*24*5)),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	return &c, nil
}

func AValidCourseWithIDs(ctx context.Context, eurekaDB database.Ext, studentIDs []string, courseID string) (courseStudents []*entities.CourseStudent, _ error) {
	for i := 0; i < len(studentIDs); i++ {
		courseStudent, err := generateCourseByStudentID(studentIDs[i], courseID)
		if err != nil {
			return nil, fmt.Errorf("generateCourseByStudentID: %w", err)
		}
		courseStudents = append(courseStudents, courseStudent)

		cmd, err := database.Insert(ctx, courseStudent, eurekaDB.Exec)
		if err != nil {
			return nil, fmt.Errorf("database.Insert: %w", err)
		}
		if cmd.RowsAffected() != 1 {
			return nil, fmt.Errorf("error insert course student")
		}
	}
	return courseStudents, nil
}

type GenerateAssignmentResult struct{ AssignmentIDs []string }

func GenerateAssignment(ctx context.Context, topicID string, numberAssigment int, loIDs []string, eurekaConn *grpc.ClientConn, template *epb.Assignment) (*GenerateAssignmentResult, error) {
	pbAssignments := make([]*epb.Assignment, 0)

	for i := 0; i < numberAssigment; i++ {
		assignment := template
		if template == nil {
			// Default assignment template
			assignmentID := idutil.ULIDNow()
			assignment = &epb.Assignment{
				AssignmentId: assignmentID,
				Name:         fmt.Sprintf("assignment-%s", assignmentID),
				Content: &epb.AssignmentContent{
					TopicId: topicID,
					LoId:    loIDs,
				},
				CheckList: &epb.CheckList{
					Items: []*epb.CheckListItem{
						{
							Content:   "Complete all learning objectives",
							IsChecked: true,
						},
						{
							Content:   "Submitted required videos",
							IsChecked: false,
						},
					},
				},
				Instruction:    "teacher's instruction",
				MaxGrade:       10,
				Attachments:    []string{"media-id-1", "media-id-2"},
				AssignmentType: epb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE,
				Setting: &epb.AssignmentSetting{
					AllowLateSubmission: true,
					AllowResubmission:   true,
				},
				RequiredGrade: true,
				DisplayOrder:  0,
			}
		}

		pbAssignments = append(pbAssignments, assignment)
	}
	resp, err := epb.NewAssignmentModifierServiceClient(eurekaConn).UpsertAssignments(ctx, &epb.UpsertAssignmentsRequest{
		Assignments: pbAssignments,
	})

	if err != nil {
		return nil, fmt.Errorf("unable create a assignment: %v", err)
	}
	if resp.GetAssignmentIds() == nil {
		return nil, fmt.Errorf("error AssignmentId is nil")
	}
	result := &GenerateAssignmentResult{AssignmentIDs: resp.GetAssignmentIds()}

	return result, nil
}

func GenerateCourseBooks(ctx context.Context, courseID string, bookIDs []string, eurekaConn *grpc.ClientConn) error {
	if _, err := epb.NewCourseModifierServiceClient(eurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: courseID,
		BookIds:  bookIDs,
	},
	); err != nil {
		return fmt.Errorf("unable to add book to course: %w", err)
	}
	return nil
}

func GenerateQuizzes(ctx context.Context, loID string, numberOfQuizzes int, template []*epb.QuizLO, eurekaConn *grpc.ClientConn) error {
	reqs := make([]*sspb.UpsertFlashcardContentRequest, 0)

	for i := 0; i < numberOfQuizzes; i++ {
		quizzes := template
		if template == nil {
			quizzes = []*epb.QuizLO{
				{
					Quiz: &cpb.QuizCore{
						ExternalId: idutil.ULIDNow(),
						Kind:       cpb.QuizType_QUIZ_TYPE_FIB,
						Info: &cpb.ContentBasicInfo{
							SchoolId: constants.ManabieSchool,
							Country:  cpb.Country_COUNTRY_VN,
						},
						Question: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Explanation: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						TaggedLos:       []string{"123", "abc"},
						DifficultyLevel: 2,
						Options: []*cpb.QuizOption{
							{
								Content: &cpb.RichText{
									Raw: `
									{
										"blocks": [
											{
												"key": "2lnf5",
												"text": "A",
												"type": "unstyled",
												"depth": 0,
												"inlineStyleRanges": [],
												"entityRanges": [],
												"data": {}
											}
										],
										"entityMap": {}
									}
								`,
									Rendered: "rendered " + idutil.ULIDNow(),
								},
								Attribute: &cpb.QuizItemAttribute{
									Configs: []cpb.QuizItemAttributeConfig{
										cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
									},
								},
								Correctness: true,
								Label:       "(1)",
								Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
								Key:         idutil.ULIDNow(),
							},
						},
						Attribute: &cpb.QuizItemAttribute{
							Configs: []cpb.QuizItemAttributeConfig{
								cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
							},
						},
						Point: wrapperspb.Int32(1),
					},
					LoId: loID,
				},
			}
		}
		quizCores := make([]*cpb.QuizCore, 0)

		for iter := 0; iter < len(quizzes); iter++ {
			quizCores = append(quizCores, quizzes[iter].Quiz)
		}
		reqs = append(reqs, &sspb.UpsertFlashcardContentRequest{Quizzes: quizCores, FlashcardId: loID})
	}
	for _, req := range reqs {
		_, err := sspb.NewQuizClient(eurekaConn).UpsertFlashcardContent(ctx, req)
		if err != nil {
			return fmt.Errorf("generate quizzes error: %w", err)
		}
	}

	return nil
}

func GenerateQuizzesWithTag(ctx context.Context, loID string, numberOfQuizzes int, tags [][]string, eurekaConn *grpc.ClientConn) (quizIDs []string, err error) {
	reqs := make([]*sspb.UpsertFlashcardContentRequest, 0)
	if len(tags) != numberOfQuizzes {
		return nil, fmt.Errorf("length of tags is not equal to number of quizzes")
	}
	quizIDs = make([]string, 0)
	for i := 0; i < numberOfQuizzes; i++ {
		quizExternalID := idutil.ULIDNow()
		quizIDs = append(quizIDs, quizExternalID)
		quiz := &epb.QuizLO{
			Quiz: &cpb.QuizCore{
				ExternalId: quizExternalID,
				Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
				Info: &cpb.ContentBasicInfo{
					SchoolId: constants.ManabieSchool,
					Country:  cpb.Country_COUNTRY_VN,
				},
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				QuestionTagIds:  tags[i],
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw: `
									{
										"blocks": [
											{
												"key": "2lnf5",
												"twt": "A",
												"type": "unstyled",
												"depth": 0,
												"inlineStyleRanges": [],
												"entityRanges": [],
												"data": {}
											}
										],
										"entityMap": {}
									}
								`,
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Attribute: &cpb.QuizItemAttribute{
							Configs: []cpb.QuizItemAttributeConfig{
								cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
							},
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
				},
				Attribute: &cpb.QuizItemAttribute{
					Configs: []cpb.QuizItemAttributeConfig{
						cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
					},
				},
				Point: wrapperspb.Int32(1),
			},
			LoId: loID,
		}

		reqs = append(reqs, &sspb.UpsertFlashcardContentRequest{Quizzes: []*cpb.QuizCore{quiz.Quiz}, FlashcardId: loID})
	}
	for _, req := range reqs {
		_, err := sspb.NewQuizClient(eurekaConn).UpsertFlashcardContent(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("generate quizzes error: %w", err)
		}
	}

	return quizIDs, nil
}

func UserInsertALearningObjective(ctx context.Context, eurekaConn *grpc.ClientConn, topicID string) (learningMaterialId string, err error) {
	lo := &sspb.LearningObjectiveBase{
		Base: &sspb.LearningMaterialBase{
			TopicId: topicID,
			Name:    "LearningObjective",
			Type:    sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String(),
		},
	}
	response, err := sspb.NewLearningObjectiveClient(eurekaConn).InsertLearningObjective(ctx, &sspb.InsertLearningObjectiveRequest{
		LearningObjective: lo,
	})
	if err != nil {
		return response.LearningMaterialId, err
	}
	return response.LearningMaterialId, nil
}

func UserAssignCourseToAStudent(ctx context.Context, userMgmtConn *grpc.ClientConn, studentID, courseID string) (err error) {
	_, err = ys_pb_v1.NewUserModifierServiceClient(userMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
		StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: courseID,
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
		StudentId: studentID,
	})
	if err != nil {
		return fmt.Errorf("UpsertStudentCoursePackage: %s", err.Error())
	}
	return nil
}

func GenerateStudyPlans(ctx context.Context, eurekaConn *grpc.ClientConn, courseID, bookID string, numberOfStudyPlan int) (studyPlanID []string, err error) {
	if resp, err := epb.NewCourseModifierServiceClient(eurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: courseID,
		BookIds:  []string{bookID},
	}); err != nil || !resp.Successful {
		err = fmt.Errorf("NewCourseModifierServiceClient.AddBooks: %w", err)
		return nil, err
	}
	studyPlanIDs := make([]string, 0, numberOfStudyPlan)
	for i := 0; i < numberOfStudyPlan; i++ {
		resp, err := epb.NewStudyPlanModifierServiceClient(eurekaConn).UpsertStudyPlan(ctx, &epb.UpsertStudyPlanRequest{
			Name:                fmt.Sprintf("studyplan-%s", bookID),
			SchoolId:            constants.ManabieSchool,
			TrackSchoolProgress: true,
			Grades:              []int32{3, 4},
			Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
			BookId:              bookID,
			CourseId:            courseID,
		})
		if err != nil {
			return nil, fmt.Errorf("NewStudyPlanModifierServiceClient.UpsertStudyPlan: %w", err)
		}
		studyPlanIDs = append(studyPlanIDs, resp.StudyPlanId)
	}
	return studyPlanIDs, nil
}

func GenerateQuestionTags(ctx context.Context, eurekaDB *pgxpool.Pool, questionTagIDs []string, questionTagTypeID string) error {
	stmt := `INSERT INTO question_tag(question_tag_id, name, question_tag_type_id, created_at, updated_at, deleted_at) VALUES %s`
	value := ""
	for _, questionTagID := range questionTagIDs {
		name := fmt.Sprintf("tag-name-%s", questionTagID)
		value = value + "('" + questionTagID + "', '" + name + "', '" + questionTagTypeID + "', NOW() , NOW(), NULL), "
	}
	_, err := eurekaDB.Exec(ctx, fmt.Sprintf(stmt, value[:len(value)-2]))
	if err != nil {
		return fmt.Errorf("cannot create question tags: %s", err.Error())
	}
	return nil
}

func GenerateQuestionTagType(ctx context.Context, eurekaDB *pgxpool.Pool, questionTagTypeID string) error {
	stmt := `INSERT INTO question_tag_type(question_tag_type_id, name, created_at, updated_at, deleted_at) VALUES %s`
	value := "('" + questionTagTypeID + "', '" + idutil.ULIDNow() + "', NOW() , NOW(), NULL), "

	_, err := eurekaDB.Exec(ctx, fmt.Sprintf(stmt, value[:len(value)-2]))
	if err != nil {
		return fmt.Errorf("cannot create question tag type: %s", err.Error())
	}
	return nil
}

func GenerateExamLO(ctx context.Context, topicID string, template *sspb.ExamLOBase, gradeToPass *wrapperspb.Int32Value, manualGrading bool, approveGrading bool, eurekaConn *grpc.ClientConn) (learningMaterialID string, err error) {
	lo := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:        idutil.ULIDNow(),
			Name:      fmt.Sprintf("exam-lo-%s", topicID),
			Country:   cpb.Country_COUNTRY_VN,
			Grade:     1,
			Subject:   cpb.Subject_SUBJECT_MATHS,
			MasterId:  "",
			SchoolId:  constants.ManabieSchool,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
		TopicId: topicID,
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide:     "https://guides/1/master",
		Video:          "https://videos/1/master",
		Type:           cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_EXAM_LO,
		GradeToPass:    gradeToPass,
		ManualGrading:  manualGrading,
		ApproveGrading: approveGrading,
	}

	resp, err := epb.NewLearningObjectiveModifierServiceClient(eurekaConn).UpsertLOs(ctx, &epb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			lo,
		},
	})
	if err != nil {
		return "", fmt.Errorf("NewLearningObjectiveModifierServiceClient.UpsertLOs: %w", err)
	}

	return resp.LoIds[0], nil
}

type GenerateUpsertSingleQuizResult struct {
	ExternalIDs []string
	TotalPoint  int32
}

func GenerateUpsertSingleQuiz(ctx context.Context, loID string, quizKind cpb.QuizType, numOfQuizzes int, point int32, eurekaConn *grpc.ClientConn) (*GenerateUpsertSingleQuizResult, error) {
	externalIDs := make([]string, 0, numOfQuizzes)
	var totalPoint int32

	for i := 0; i < numOfQuizzes; i++ {
		var quizLO *epb.QuizLO

		switch quizKind {
		case cpb.QuizType_QUIZ_TYPE_MCQ:
			quizLO = GetQuizTypeMCQTemplate(loID, point)
		case cpb.QuizType_QUIZ_TYPE_MAQ:
			quizLO = GetQuizTypeMAQTemplate(loID, point)
		case cpb.QuizType_QUIZ_TYPE_MIQ:
			quizLO = GetQuizTypeMIQTemplate(loID, point)
		case cpb.QuizType_QUIZ_TYPE_FIB:
			quizLO = GetQuizTypeFIBTemplate(loID, point)
		case cpb.QuizType_QUIZ_TYPE_ORD:
			quizLO = GetQuizTypeORDTemplate(loID, point)
		}

		if _, err := epb.NewQuizModifierServiceClient(eurekaConn).UpsertSingleQuiz(ctx, &epb.UpsertSingleQuizRequest{
			QuizLo: quizLO,
		}); err != nil {
			return nil, fmt.Errorf("NewQuizModifierServiceClient.UpsertSingleQuiz: %w", err)
		}

		externalIDs = append(externalIDs, quizLO.Quiz.ExternalId)
		totalPoint += quizLO.Quiz.Point.Value
	}

	return &GenerateUpsertSingleQuizResult{ExternalIDs: externalIDs, TotalPoint: totalPoint}, nil
}

func GenerateUpsertFlashcardContent(ctx context.Context, loID string, point int32, eurekaConn *grpc.ClientConn) ([]string, error) {
	quizLOs := GetQuizTypePOWTemplate(loID, point)
	quizzes := make([]*cpb.QuizCore, 0)
	quizIDs := make([]string, 0)
	for _, quizLO := range quizLOs {
		quizIDs = append(quizIDs, quizLO.Quiz.ExternalId)
		quizzes = append(quizzes, quizLO.Quiz)
	}

	if _, err := sspb.NewQuizClient(eurekaConn).UpsertFlashcardContent(ctx, &sspb.UpsertFlashcardContentRequest{
		FlashcardId: loID,
		Quizzes:     quizzes,
	}); err != nil {
		return nil, fmt.Errorf("NewQuizClient.UpsertFlashcardContent: %w", err)
	}

	return quizIDs, nil
}

func GetQuizTypeMCQTemplate(loID string, point int32) *epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	// Question: Question 1
	quizRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Question 1",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizRaw, _ := json.Marshal(quizRawObj)
	// Answer: "A" => is correct
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)
	// Answer: "B"
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)
	// Answer: "C"
	answerCRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "C",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerCRaw, _ := json.Marshal(answerCRawObj)

	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_MCQ,
			Question: &cpb.RichText{
				Raw: string(quizRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerCRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
			},
			Config:    []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{},
			Point:     wrapperspb.Int32(point),
		},
		LoId: loID,
	}

	return quizLO
}

func GetQuizTypeMAQTemplate(loID string, point int32) *epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	// Question: Question 1
	quizRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Question 1",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizRaw, _ := json.Marshal(quizRawObj)
	// Answer: "A" => is correct
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)
	// Answer: "B"
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)
	// Answer: "C"
	answerCRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "C",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerCRaw, _ := json.Marshal(answerCRawObj)

	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_MAQ,
			Question: &cpb.RichText{
				Raw: string(quizRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
					},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
					},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
					},
					Content: &cpb.RichText{
						Raw: string(answerCRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
			},
			Config:    []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{},
			Point:     wrapperspb.Int32(point),
		},
		LoId: loID,
	}

	return quizLO
}

func GetQuizTypeMIQTemplate(loID string, point int32) *epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	// Question: Question 1
	quizRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Question 1",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizRaw, _ := json.Marshal(quizRawObj)
	// Answer: Correct
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)
	// Answer: Incorrect
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)

	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_MIQ,
			Question: &cpb.RichText{
				Raw: string(quizRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
			},
			Config:    []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{},
			Point:     wrapperspb.Int32(point),
		},
		LoId: loID,
	}

	return quizLO
}

func GetQuizTypeFIBTemplate(loID string, point int32) *epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	// Question: Question 1
	quizRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Question 1",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizRaw, _ := json.Marshal(quizRawObj)
	// Answer: "A" => is correct
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)
	// Answer: "B"
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)
	// Answer: "C"
	answerCRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "C",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerCRaw, _ := json.Marshal(answerCRawObj)

	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_FIB,
			Question: &cpb.RichText{
				Raw: string(quizRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
					},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
					},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:   idutil.ULIDNow(),
					Label: "",
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT,
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
					},
					Content: &cpb.RichText{
						Raw: string(answerCRaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
			},
			Config:    []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{},
			Point:     wrapperspb.Int32(point),
		},
		LoId: loID,
	}

	return quizLO
}

func InsertANewQuestionGroup(ctx context.Context, eurekaConn *grpc.ClientConn, lmID string) (*sspb.UpsertQuestionGroupResponse, error) {
	req := &sspb.UpsertQuestionGroupRequest{
		LearningMaterialId: lmID,
		Name:               "name_" + lmID,
		Description:        "description_" + lmID,
		RichDescription: &cpb.RichText{
			Raw:      "raw rich text",
			Rendered: "rendered rich text",
		},
	}

	return UpsertQuestionGroup(ctx, eurekaConn, req)
}

func UpsertQuestionGroup(ctx context.Context, eurekaConn *grpc.ClientConn, req *sspb.UpsertQuestionGroupRequest) (*sspb.UpsertQuestionGroupResponse, error) {
	if len(req.LearningMaterialId) == 0 {
		return nil, fmt.Errorf("lo ID dont have yet")
	}
	return sspb.NewQuestionServiceClient(eurekaConn).
		UpsertQuestionGroup(ctx, req)
}

func GenerateQuizLOProtobufMessage(num int, loID string) []*epb.QuizLO {
	res := make([]*epb.QuizLO, 0, num)
	for i := 0; i < num; i++ {
		res = append(
			res,
			&epb.QuizLO{
				Quiz: &cpb.QuizCore{
					ExternalId: idutil.ULIDNow(),
					Kind:       cpb.QuizType_QUIZ_TYPE_POW,
					Info: &cpb.ContentBasicInfo{
						SchoolId: constants.ManabieSchool,
						Country:  cpb.Country_COUNTRY_VN,
					},
					Question: &cpb.RichText{
						Raw:      "raw",
						Rendered: "rendered " + idutil.ULIDNow(),
					},
					Explanation: &cpb.RichText{
						Raw:      "raw",
						Rendered: "rendered " + idutil.ULIDNow(),
					},
					TaggedLos:       []string{"123", "abc"},
					DifficultyLevel: 2,
					Options: []*cpb.QuizOption{
						{
							Content: &cpb.RichText{
								Raw:      "raw",
								Rendered: "rendered " + idutil.ULIDNow(),
							},
							Correctness: false,
							Configs: []cpb.QuizOptionConfig{
								cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
							},
							Attribute: &cpb.QuizItemAttribute{
								ImgLink:   "img.link",
								AudioLink: "audio.link",
								Configs: []cpb.QuizItemAttributeConfig{
									1,
								},
							},
							Label: "label",
							Key:   "key",
						},
					},
					Attribute: &cpb.QuizItemAttribute{
						ImgLink:   "img.link",
						AudioLink: "audio.link",
						Configs: []cpb.QuizItemAttributeConfig{
							1,
						},
					},
					Point: wrapperspb.Int32(rand.Int31n(10)),
				},
				LoId: loID,
			})
	}
	return res
}

func UpsertQuizzes(ctx context.Context, eurekaConn *grpc.ClientConn, req []*epb.QuizLO) ([]*epb.UpsertSingleQuizResponse, error) {
	res := make([]*epb.UpsertSingleQuizResponse, 0, len(req))
	for _, item := range req {
		r, err := epb.NewQuizModifierServiceClient(eurekaConn).UpsertSingleQuiz(ctx, &epb.UpsertSingleQuizRequest{
			QuizLo: item,
		})
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	return res, nil
}

func UserAssignStudyPlanToAStudent(ctx context.Context, eurekaConn *grpc.ClientConn, studentID, studyPlanID string) (err error) {
	req := &epb.AssignStudyPlanRequest{
		StudyPlanId: studyPlanID,
		Data: &epb.AssignStudyPlanRequest_StudentId{
			StudentId: studentID,
		},
	}
	_, err = epb.NewAssignmentModifierServiceClient(eurekaConn).AssignStudyPlan(ctx, req)

	return err
}

func GenerateIndividualStudyPlanRequest(spID, lmID, studentID string) *sspb.UpsertIndividualInfoRequest {
	req := &sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        spID,
					LearningMaterialId: lmID,
					StudentId: &wrapperspb.StringValue{
						Value: studentID,
					},
				},
				AvailableFrom: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:   timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:     timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:       timestamppb.New(time.Now().AddDate(0, 0, 1)),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	}

	return req
}

func GetQuizTypePOWTemplate(loID string, point int32) []*epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	quizLOs := make([]*epb.QuizLO, 0)

	// Term A - Mean A
	quizARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Term A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizARaw, _ := json.Marshal(quizARawObj)
	// Answer: "Mean A"
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Mean A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)

	quizALO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_POW,
			Question: &cpb.RichText{
				Raw: string(quizARaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizARaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute: &cpb.QuizItemAttribute{
						Configs: []cpb.QuizItemAttributeConfig{
							cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
						},
					},
				},
			},
			Config: []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{
				Configs: []cpb.QuizItemAttributeConfig{
					cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP,
				},
			},
			Point: wrapperspb.Int32(point),
		},
		LoId: loID,
	}
	quizLOs = append(quizLOs, quizALO)

	// Term B - Mean B
	quizBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Term B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizBRaw, _ := json.Marshal(quizBRawObj)
	// Answer: "Mean B"
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Mean B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)

	quizBLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_POW,
			Question: &cpb.RichText{
				Raw: string(quizBRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizBRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:     idutil.ULIDNow(),
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: true,
					Attribute: &cpb.QuizItemAttribute{
						Configs: []cpb.QuizItemAttributeConfig{
							cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
						},
					},
				},
			},
			Config: []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{
				Configs: []cpb.QuizItemAttributeConfig{
					cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP,
				},
			},
			Point: wrapperspb.Int32(point),
		},
		LoId: loID,
	}
	quizLOs = append(quizLOs, quizBLO)

	return quizLOs
}

func GetQuizTypeORDTemplate(loID string, point int32) *epb.QuizLO {
	rand.Seed(time.Now().UnixNano())

	// Question: Question 1
	quizRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "Question 1",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	quizRaw, _ := json.Marshal(quizRawObj)
	// Answer: "A" => is correct
	answerARawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "A",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerARaw, _ := json.Marshal(answerARawObj)
	// Answer: "B"
	answerBRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "B",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerBRaw, _ := json.Marshal(answerBRawObj)
	// Answer: "C"
	answerCRawObj := Raw{
		Blocks: []Block{
			{
				Key:               idutil.ULIDNow(),
				Text:              "C",
				Type:              "unstyled",
				Depth:             0,
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              struct{}{},
			},
		},
		EntityMap: struct{}{},
	}
	answerCRaw, _ := json.Marshal(answerCRawObj)

	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			ExternalId: idutil.ULIDNow(),
			Kind:       cpb.QuizType_QUIZ_TYPE_ORD,
			Question: &cpb.RichText{
				Raw: string(quizRaw),
			},
			Explanation: &cpb.RichText{
				Raw: string(quizRaw),
			},
			DifficultyLevel: rand.Int31n(5) + 1,
			TaggedLos:       []string{loID},
			Options: []*cpb.QuizOption{
				{
					Key:     "keyA",
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerARaw),
					},
					Correctness: true,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:     "keyB",
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerBRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
				{
					Key:     "keyC",
					Label:   "",
					Configs: []cpb.QuizOptionConfig{},
					Content: &cpb.RichText{
						Raw: string(answerCRaw),
					},
					Correctness: false,
					Attribute:   &cpb.QuizItemAttribute{},
				},
			},
			Config:    []cpb.QuizConfig{},
			Attribute: &cpb.QuizItemAttribute{},
			Point:     wrapperspb.Int32(point),
		},
		LoId: loID,
	}

	return quizLO
}

func CryptRand(max int64) int32 {
	n, err := crypRand.Int(crypRand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return int32(n.Int64())
}
