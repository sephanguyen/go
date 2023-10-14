package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateBookContentInput struct {
	BookName    string
	ChapterName string
	TopicName   string
}

type CreateBookContentOutput struct {
	Book    *entities.Book
	Chapter *entities.Chapter
	Topic   *entities.Topic
}

type CreateAdHocTaskAssignmentInput struct {
	TopicID string
	Name    string
}

type CreateStudyPlanInput struct {
	StudyPlanName      string
	BookID             string
	CourseID           string
	LearningMaterialID string
	StarDate           *timestamppb.Timestamp
	EndDate            *timestamppb.Timestamp
}

type IStudentAdHocFactory interface {
	SetStudent(*bpb.StudentProfile) error
	CreateAdHocBookContent(context.Context, CreateBookContentInput) (*CreateBookContentOutput, error)
	CreateAdHocStudyPlan(context.Context, CreateStudyPlanInput) error
}

type StudentAdHocFactory struct {
	DB      database.Ext
	Student *bpb.StudentProfile
	Grade   int

	BookRepo                    IBookRepository
	ChapterRepo                 IChapterRepository
	BookChapterRepo             IBookChapterRepository
	TopicRepo                   ITopicRepository
	CourseBookRepo              ICourseBookRepository
	StudyPlanRepo               IStudyPlanRepository
	StudentStudyPlanRepo        IStudentStudyPlanRepository
	AssignmentRepo              IAssignmentRepository
	StudyPlanItemRepo           IStudyPlanItemRepository
	AssignmentStudyPlanItemRepo IAssignmentStudyPlanItemRepository
	LoStudyPlanItemRepo         ILoStudyPlanItemRepository
	LearningObjectiveRepo       ILearningObjectiveRepository
	MasterStudyPlanRepo         IMasterStudyPlanRepository
	IndividualStudyPlanRepo     IIndividualStudyPlanRepository
}

func NewStudentAdHocFactory(
	db database.Ext,
	bookRepo IBookRepository,
	chapterRepo IChapterRepository,
	bookChapterRepo IBookChapterRepository,
	topicRepo ITopicRepository,
	courseBookRepo ICourseBookRepository,
	studyPlanRepo IStudyPlanRepository,
	studentStudyPlanRepo IStudentStudyPlanRepository,
	assignmentRepo IAssignmentRepository,
	studyPlanItemRepo IStudyPlanItemRepository,
	assignmentStudyPlanItemRepo IAssignmentStudyPlanItemRepository,
	loStudyPlanItemRepo ILoStudyPlanItemRepository,
	learningObjectiveRepo ILearningObjectiveRepository,
	masterStudyPlanRepo IMasterStudyPlanRepository,
	individualStudyPlanRepo IIndividualStudyPlanRepository,
) IStudentAdHocFactory {
	return &StudentAdHocFactory{
		DB:                          db,
		BookRepo:                    bookRepo,
		ChapterRepo:                 chapterRepo,
		BookChapterRepo:             bookChapterRepo,
		TopicRepo:                   topicRepo,
		CourseBookRepo:              courseBookRepo,
		StudyPlanRepo:               studyPlanRepo,
		StudentStudyPlanRepo:        studentStudyPlanRepo,
		AssignmentRepo:              assignmentRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		LearningObjectiveRepo:       learningObjectiveRepo,
		MasterStudyPlanRepo:         masterStudyPlanRepo,
		IndividualStudyPlanRepo:     individualStudyPlanRepo,
	}
}

func (s *StudentAdHocFactory) SetStudent(student *bpb.StudentProfile) (err error) {
	s.Student = student
	if s.Grade, err = i18n.ConvertStringGradeToInt(bob_pb.Country(student.Country), student.Grade); err != nil {
		return fmt.Errorf("i18n.ConvertStringGradeToInt: %w", err)
	}
	return
}

func (s *StudentAdHocFactory) CreateAdHocBookContent(ctx context.Context, input CreateBookContentInput) (*CreateBookContentOutput, error) {
	if s.Student == nil {
		return nil, fmt.Errorf("student profile must not be nil")
	}
	// create book
	if input.BookName == "" {
		input.BookName = s.Student.Name + "'s Todo"
	}
	bookID := idutil.ULIDNow()
	book := &entities.Book{}
	database.AllNullEntity(book)
	if err := multierr.Combine(
		book.ID.Set(bookID),
		book.Name.Set(input.BookName),
		book.SchoolID.Set(s.Student.School.Id),
		book.Country.Set(s.Student.Country.String()),
		book.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		book.Grade.Set(s.Grade),
		book.BookType.Set(cpb.BookType_BOOK_TYPE_ADHOC.String()),
		book.CurrentChapterDisplayOrder.Set(0),
	); err != nil {
		return nil, fmt.Errorf("can not set book data: %w", err)
	}
	if err := s.BookRepo.Upsert(ctx, s.DB, []*entities.Book{book}); err != nil {
		return nil, fmt.Errorf("s.BookRepo.Upsert: %w", err)
	}

	// create chapter
	if input.ChapterName == "" {
		input.ChapterName = s.Student.Name + "'s Todo"
	}
	chapterID := idutil.ULIDNow()
	chapter := &entities.Chapter{}
	database.AllNullEntity(chapter)
	if err := multierr.Combine(
		chapter.ID.Set(chapterID),
		chapter.Name.Set(input.ChapterName),
		chapter.SchoolID.Set(s.Student.School.Id),
		chapter.Country.Set(s.Student.Country.String()),
		chapter.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		chapter.Grade.Set(s.Grade),
		chapter.DisplayOrder.Set(0),
	); err != nil {
		return nil, fmt.Errorf("can not set chapter data: %w", err)
	}
	if err := s.ChapterRepo.Upsert(ctx, s.DB, []*entities.Chapter{chapter}); err != nil {
		return nil, fmt.Errorf("s.ChapterRepo.Upsert: %w", err)
	}

	// create book_chapter
	bookChapter := &entities.BookChapter{}
	database.AllNullEntity(bookChapter)
	if err := multierr.Combine(
		bookChapter.BookID.Set(bookID),
		bookChapter.ChapterID.Set(chapterID),
	); err != nil {
		return nil, fmt.Errorf("can not set book_chapter data: %w", err)
	}
	if err := s.BookChapterRepo.Upsert(ctx, s.DB, []*entities.BookChapter{bookChapter}); err != nil {
		return nil, fmt.Errorf("s.BookChapterRepo.Upsert: %w", err)
	}

	// create topic
	if input.TopicName == "" {
		input.TopicName = s.Student.Name + "'s Todo"
	}
	topicID := idutil.ULIDNow()
	topic := &entities.Topic{}
	database.AllNullEntity(topic)
	if err := multierr.Combine(
		topic.ID.Set(topicID),
		topic.Name.Set(input.TopicName),
		topic.SchoolID.Set(s.Student.School.Id),
		topic.Country.Set(s.Student.Country.String()),
		topic.Grade.Set(s.Grade),
		topic.Subject.Set(cpb.Subject_SUBJECT_NONE.String()),
		topic.TopicType.Set(pb.TopicType_TOPIC_TYPE_NONE.String()),
		topic.Status.Set(pb.TopicStatus_TOPIC_STATUS_NONE.String()),
		topic.DisplayOrder.Set(0),
		topic.ChapterID.Set(chapterID),
		topic.TotalLOs.Set(0),
		topic.EssayRequired.Set(false),
	); err != nil {
		return nil, fmt.Errorf("can not set topic data: %w", err)
	}
	if err := s.TopicRepo.BulkImport(ctx, s.DB, []*entities.Topic{topic}); err != nil {
		return nil, fmt.Errorf("s.TopicRepo.BulkImport: %w", err)
	}
	return &CreateBookContentOutput{
		Book:    book,
		Chapter: chapter,
		Topic:   topic,
	}, nil
}

func (s *StudentAdHocFactory) CreateAdHocStudyPlan(ctx context.Context, input CreateStudyPlanInput) error {
	if s.Student == nil {
		return fmt.Errorf("student profile must not be nil")
	}
	// create course_book
	courseBook := &entities.CoursesBooks{}
	database.AllNullEntity(courseBook)
	if err := multierr.Combine(
		courseBook.BookID.Set(input.BookID),
		courseBook.CourseID.Set(input.CourseID),
	); err != nil {
		return fmt.Errorf("can not set course_book data: %w", err)
	}
	if err := s.CourseBookRepo.Upsert(ctx, s.DB, []*entities.CoursesBooks{courseBook}); err != nil {
		return fmt.Errorf("s.CourseBookRepo.Upsert: %w", err)
	}

	if input.StudyPlanName == "" {
		input.StudyPlanName = s.Student.Name + "'s Todo"
	}

	studyPlan, err := convertUpsertAdHocIndividualStudyPlanRequestToStudyPlanEntity(&pb.UpsertAdHocIndividualStudyPlanRequest{
		Name:      input.StudyPlanName,
		BookId:    input.BookID,
		CourseId:  input.CourseID,
		StudentId: s.Student.Id,
	})

	if err != nil {
		return fmt.Errorf("s.CreateAdHocStudyPlan.convertUpsertAdHocIndividualStudyPlanRequestToStudyPlanEntity: %w", err)
	}
	studyPlans := []*entities.StudyPlan{studyPlan}
	if err := s.StudyPlanRepo.BulkUpsert(ctx, s.DB, studyPlans); err != nil {
		return fmt.Errorf("studyPlanRepo.BulkUpsert: %w", err)
	}

	ssp, err := toStudentStudyPlanEn(database.Text(s.Student.Id), studyPlan.ID.String)
	if err != nil {
		return fmt.Errorf("toStudentStudyPlanEn: %w", err)
	}
	if err := s.StudentStudyPlanRepo.BulkUpsert(ctx, s.DB, []*entities.StudentStudyPlan{ssp}); err != nil {
		return fmt.Errorf("studentStudyPlan.BulkUpsert: %w", err)
	}

	t, _ := time.Parse("2006/02/01 15:04", "2300/01/01 23:59")

	masterStudyPlan, err := toMasterStudyPlanEntity(&sspb.MasterStudyPlan{
		MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
			StudyPlanId:        studyPlan.ID.String,
			LearningMaterialId: input.LearningMaterialID,
		},
		AvailableFrom: input.StarDate,
		AvailableTo:   timestamppb.New(t),
		StartDate:     input.StarDate,
		EndDate:       input.EndDate,
		Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
	})
	if err != nil {
		return fmt.Errorf("toMasterStudyPlanEntity: %w", err)
	}
	if err := s.MasterStudyPlanRepo.BulkUpsert(ctx, s.DB, []*entities.MasterStudyPlan{masterStudyPlan}); err != nil {
		return fmt.Errorf("s.MasterStudyPlanRepo.BulkUpsert: %w", err)
	}

	individualStudyPlans, err := toIndividualStudyPlansEnt(&sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        studyPlan.ID.String,
					LearningMaterialId: input.LearningMaterialID,
				},
				AvailableFrom: input.StarDate,
				AvailableTo:   timestamppb.New(t),
				StartDate:     input.StarDate,
				EndDate:       input.EndDate,
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("toIndividualStudyPlansEnt: %w", err)
	}
	if _, err := s.IndividualStudyPlanRepo.BulkSync(ctx, s.DB, individualStudyPlans); err != nil {
		return fmt.Errorf("s.MasterStudyPlanRepo.BulkUpsert: %w", err)
	}
	return nil
}
