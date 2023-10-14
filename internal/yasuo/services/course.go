package services

import (
	"context"
	"fmt"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	brightcove_service "github.com/manabie-com/backend/internal/golibs/brightcove"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	"github.com/manabie-com/backend/internal/yasuo/entities"
	y_repositories "github.com/manabie-com/backend/internal/yasuo/repositories"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrMissingCourseID          = status.Error(codes.InvalidArgument, "course id cannot be empty")
	ErrMissingName              = status.Error(codes.InvalidArgument, "course name cannot be empty")
	ErrDisplayOrderLessThanZero = status.Error(codes.InvalidArgument, "display_order cannot be less than 0")
)

type EurekaInternalModifierService interface {
	DeleteLOStudyPlanItems(ctx context.Context, req *epb.DeleteLOStudyPlanItemsRequest, opts ...grpc.CallOption) (*epb.DeleteLOStudyPlanItemsResponse, error)
}

// BobMediaModifierService is implemented by bpb.MediaModifierServiceClient.
type BobMediaModifierServiceClient interface {
	GenerateAudioFile(ctx context.Context, in *bpb.GenerateAudioFileRequest, opts ...grpc.CallOption) (*bpb.GenerateAudioFileResponse, error)
}

type CourseService struct {
	Env                       string
	EurekaDBTrace             database.Ext
	DBTrace                   database.Ext
	LessonDBTrace             database.Ext
	UnleashClientIns          unleashclient.ClientInstance
	JSM                       nats.JetStreamManagement
	Logger                    *zap.Logger
	Config                    *configurations.Config
	TopicQuestionPublish      string
	SubQuestionRenderFinish   string
	LimitQuestionsPullPerTime int
	MaxWaitTimePullQuestion   time.Duration

	BrightcoveExtService brightcove_service.ExternalService
	BrightcoveService    BrightcoveService

	BrightCoveProfile string
	UserRepo          interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.User, error)
	}
	ChapterRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities_bob.Chapter) error
		FindSchoolIDsOnChapters(ctx context.Context, db database.QueryExecer, chapterIDs []string) ([]int32, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs []string) (int, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities_bob.Chapter, error)
		UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, cc []*entities_bob.Chapter) error
		FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities_bob.Chapter, error)
		UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error
	}
	CourseRepo interface {
		FindSchoolIDsOnCourses(ctx context.Context, db database.QueryExecer, courseIDs []string) ([]int32, error)
		FindByIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities_bob.Course, error)
		FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities_bob.Course, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error
		Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.Course) error
		UpsertV2(ctx context.Context, db database.Ext, cc []*entities_bob.Course) error
		UpdateAcademicYear(ctx context.Context, db database.Ext, cc []*repositories.UpdateAcademicYearOpts) error
	}
	CourseAccessPathRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.CourseAccessPath) error
	}
	CourseClassRepo interface {
		FindByCourseID(ctx context.Context, db database.Ext, courseID pgtype.Text, isAll bool) (map[pgtype.Int4]*entities_bob.CourseClass, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error
		UpsertV2(ctx context.Context, db database.Ext, cc []*entities_bob.CourseClass) error
		FindByClassIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapByClassID map[pgtype.Int4]pgtype.TextArray, err error)
		FindByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray, isAll bool) ([]*entities_bob.CourseClass, error)
	}
	ClassRepo interface {
		RetrieveClassIDsBySchoolIDs(ctx context.Context, tx database.QueryExecer, schoolIDs []int) (map[int][]int, error)
		FindBySchoolAndID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Int4, classIDs pgtype.Int4Array) (map[pgtype.Int4]*entities_bob.Class, error)
	}
	QuestionSetRepo interface {
		CreateAll(ctx context.Context, db database.QueryExecer, quizsets []*entities.Quizsets) error
	}
	TeacherRepo interface {
		GetTeacherHasSchoolIDs(ctx context.Context, db database.QueryExecer, teacherID string, schoolIds []int32) (*entities_bob.Teacher, error)
		IsInSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) (bool, error)
		ManyTeacherIsInSchool(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray, schoolID pgtype.Int4) (bool, error)
	}

	ActivityLogRepo interface {
		BulkCreate(ctx context.Context, db database.Ext, logs []*entities_bob.ActivityLog) error
		CreateV2(ctx context.Context, db database.Ext, log *entities_bob.ActivityLog) error
	}

	TopicRepo interface {
		FindSchoolIDs(ctx context.Context, db database.QueryExecer, topicIDs []string) ([]int32, error)
		Create(ctx context.Context, db database.Ext, plans []*entities_bob.Topic) error
		FindByID(ctx context.Context, db database.QueryExecer, ID pgtype.Text) (*entities_bob.Topic, error)
		Update(ctx context.Context, db database.QueryExecer, src *entities_bob.Topic) error
		SoftDeleteV2(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error
		SoftDeleteV3(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (int, error)
		FindByIDsV2(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities_bob.Topic, error)
	}
	LoRepo interface {
		FindSchoolIDs(context.Context, database.QueryExecer, []string) ([]int32, error)
		SoftDeleteWithLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (int64, error)
	}
	PresetStudyPlanRepo interface {
		Upsert(ctx context.Context, db database.Ext, preset *entities_bob.PresetStudyPlan) error
		SoftDelete(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray) error
	}
	PresetStudyPlanWeeklyRepo interface {
		Create(ctx context.Context, db database.Ext, plans []*entities_bob.PresetStudyPlanWeekly) error
		FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*entities_bob.PresetStudyPlanWeekly, error)
		Update(ctx context.Context, db database.QueryExecer, l *entities_bob.PresetStudyPlanWeekly) error
		SoftDelete(ctx context.Context, db database.QueryExecer, PresetStudyPlanWeeklyIDs pgtype.TextArray) error
		FindByLessonIDs(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities_bob.PresetStudyPlanWeekly, error)
		SoftDeleteByPresetStudyPlanIDs(ctx context.Context, db database.QueryExecer, PresetStudyPlanIDs pgtype.TextArray) error
	}
	LessonRepo interface {
		Create(ctx context.Context, db database.Ext, plans []*entities_bob.Lesson) error
		FindByIDs(ctx context.Context, db database.Ext, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities_bob.Lesson, error)
		FindByID(ctx context.Context, db database.Ext, ID pgtype.Text) (*entities_bob.Lesson, error)
		Update(ctx context.Context, db database.QueryExecer, src *entities_bob.Lesson) error
		SoftDelete(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error
		FindByCourseIDs(ctx context.Context, db database.Ext, IDs pgtype.TextArray, isAll bool) ([]*entities_bob.Lesson, error)
		SoftDeleteByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error
		FindEarlierAndLatestTimeLesson(ctx context.Context, db database.Ext, courseID pgtype.Text) (*time.Time, *time.Time, error)
		UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID pgtype.Text) error
		CheckExisted(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (validIDs, invalidIDs []string, err error)
		BulkUpsert(ctx context.Context, db database.Ext, lessons []*entities_bob.Lesson) error
		GetLiveLessons(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (validIDs []string, err error)
	}
	CourseBookRepo interface {
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapBookIDByCourseID map[string][]string, err error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.CoursesBooks) error
		SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs, bookIDs pgtype.TextArray) error
		FindByBookID(ctx context.Context, db database.QueryExecer, bookID string) ([]string, error)
	}
	BookRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities_bob.Book, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.Book) error
		SoftDelete(ctx context.Context, db database.QueryExecer, bookIDs []string) error
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities_bob.Book, error)
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
	}
	QuizRepo interface {
		Search(ctx context.Context,
			db database.QueryExecer, filter repositories.QuizFilter) (entities_bob.Quizzes, error)
		Create(ctx context.Context, db database.QueryExecer, quiz *entities_bob.Quiz) error
		Upsert(ctx context.Context, db database.QueryExecer, data []*entities_bob.Quiz) ([]*entities_bob.Quiz, error)
		DeleteByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) error
		GetByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) (*entities_bob.Quiz, error)
	}
	BookChapterRepo interface {
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error
		SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) error
		Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.BookChapter) error
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities_bob.BookChapter, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities_bob.BookChapter, error)
	}
	QuizSetRepo interface {
		Search(ctx context.Context,
			db database.QueryExecer, filter repositories.QuizSetFilter) (entities_bob.QuizSets, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities_bob.QuizSet) error
		Delete(ctx context.Context, db database.QueryExecer, id pgtype.Text) error
		GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities_bob.QuizSets, error)
	}
	LessonMemberRepo interface {
		UpsertQueue(b *pgx.Batch, e *entities_bob.LessonMember)
		SoftDelete(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, lessonIDs pgtype.TextArray) error
		Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities_bob.LessonMember, error)
	}
	LessonGroupRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities_bob.LessonGroup) error
		Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities_bob.LessonGroup, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities_bob.LessonGroup) error
	}
	Uploader

	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities_bob.SchoolAdmin, error)
	}
	AcademicYearRepo interface {
		Create(ctx context.Context, db database.QueryExecer, a *entities_bob.AcademicYear) error
	}
	EurekaInternalModifierService

	MediaModifierService BobMediaModifierServiceClient

	SpeechesRepo interface {
		UpsertSpeeches(ctx context.Context, db database.QueryExecer, data []*entities.Speeches) ([]*entities.Speeches, error)
		CheckExistedSpeech(ctx context.Context, db database.QueryExecer, input *y_repositories.CheckExistedSpeechReq) (bool, *entities.Speeches)
	}

	TopicLearningObjectiveRepo interface {
		SoftDeleteByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}

	LiveLessonSentNotificationRepo interface {
		SoftDeleteLiveLessonSentNotificationRecord(ctx context.Context, db database.QueryExecer, lessonID string) error
	}
}

func (s *CourseService) UpsertCourses(ctx context.Context, req *pb.UpsertCoursesRequest) (*pb.UpsertCoursesResponse, error) {
	return s.upsertCourses(ctx, req)
}

func (s *CourseService) upsertCourses(ctx context.Context, req *pb.UpsertCoursesRequest) (*pb.UpsertCoursesResponse, error) {
	courseIDs := []string{}
	upsertCourses := []*Course{}
	for _, c := range req.Courses {
		course := &Course{
			ID:           c.Id,
			Name:         c.Name,
			Country:      c.Country,
			Subject:      c.Subject,
			Grade:        c.Grade,
			DisplayOrder: c.DisplayOrder,
			ChapterIDs:   c.ChapterIds,
			SchoolID:     c.SchoolId,
			BookIDs:      c.BookIds,
			Icon:         c.Icon,
		}
		if err := validateCourseV2(course); err != nil {
			return nil, err
		}

		if c.Id != "" {
			courseIDs = append(courseIDs, c.Id)
		}

		upsertCourses = append(upsertCourses, course)
	}
	// get current courses to diff with input data
	currentCourse, err := s.CourseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(courseIDs))
	if err != nil {
		return nil, fmt.Errorf("CourseRepo.FindByIDs: %w", err)
	}

	courses := make([]*entities_bob.Course, 0, len(req.Courses))
	for _, c := range upsertCourses {
		enCourse, err := toCourseEntity(currentCourse, c)
		if err != nil {
			return nil, err
		}
		courses = append(courses, enCourse)
	}
	if err := s.CourseRepo.Upsert(ctx, s.DBTrace, courses); err != nil {
		return nil, fmt.Errorf("s.CourseRepo.Upsert: %v", err)
	}
	return &pb.UpsertCoursesResponse{Successful: true}, nil
}

type Course struct {
	ID           string
	Name         string
	Country      pb_bob.Country
	Subject      pb_bob.Subject
	Grade        string
	DisplayOrder int32
	ChapterIDs   []string
	SchoolID     int32
	Icon         string
	Status       cpb.CourseStatus
	BookIDs      []string
}

type UpsertCoursesV2Opt struct {
	Courses []*Course
}

func (s *CourseService) upsertCoursesV2(ctx context.Context, req *UpsertCoursesV2Opt) error {
	courseIDs := []string{}
	for _, c := range req.Courses {
		if err := validateCourse(c); err != nil {
			return err
		}

		if c.ID != "" {
			courseIDs = append(courseIDs, c.ID)
		}
	}

	// get current courses to diff with input data
	currentCourse, err := s.CourseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(courseIDs))
	if err != nil {
		return fmt.Errorf("CourseRepo.FindByIDs: %w", err)
	}

	courses := make([]*entities_bob.Course, 0, len(req.Courses))
	for _, c := range req.Courses {
		enCourse, err := toCourseEntity(currentCourse, c)
		if err != nil {
			return fmt.Errorf("toCourseEntity: %w", err)
		}
		courses = append(courses, enCourse)
	}

	courseAccessPaths := sliceutils.Map(courses, func(c *entities_bob.Course) *entities_bob.CourseAccessPath {
		return &entities_bob.CourseAccessPath{
			CourseID:   c.ID,
			LocationID: database.Text(constants.JPREPOrgLocation),
		}
	})

	if err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.CourseRepo.UpsertV2(ctx, tx, courses); err != nil {
			return errors.Wrap(err, "s.CourseRepo.UpsertV2")
		}
		if err := s.CourseAccessPathRepo.Upsert(ctx, tx, courseAccessPaths); err != nil {
			return errors.Wrap(err, "s.CourseAccessPathRepo.Upsert")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *CourseService) DeleteCourses(ctx context.Context, req *pb.DeleteCoursesRequest) (*pb.DeleteCoursesResponse, error) {
	if len(req.CourseIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing course id")
	}

	err := s.CourseRepo.SoftDelete(ctx, s.DBTrace, database.TextArray(req.CourseIds))
	if err != nil {
		return nil, fmt.Errorf("CourseRepo.SoftDelete: %w", err)
	}

	return &pb.DeleteCoursesResponse{Successful: true}, nil
}

func validateCourse(c *Course) error {
	if c.ID == "" {
		return status.Error(codes.InvalidArgument, "course id cannot be empty")
	}
	if c.Name == "" {
		return status.Error(codes.InvalidArgument, "course name cannot be empty")
	}
	if c.Country.String() == "" || c.Country == pb_bob.COUNTRY_NONE {
		return status.Error(codes.InvalidArgument, "country cannot be empty")
	}
	if c.Subject.String() == "" || c.Subject == pb_bob.SUBJECT_NONE {
		return status.Error(codes.InvalidArgument, "subject cannot be empty")
	}
	if c.SchoolID == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("missing school id of course %s", c.Name))
	}

	_, err := i18n.ConvertStringGradeToInt(c.Country, c.Grade)
	if err != nil {
		return err
	}
	if c.DisplayOrder < 0 {
		return status.Error(codes.InvalidArgument, "display_order cannot be less than 0")
	}

	return nil
}

// validate for upsertCourse not upsertCourseV2
func validateCourseV2(c *Course) error {
	if c.ID == "" {
		return ErrMissingCourseID
	}
	if c.Name == "" {
		return ErrMissingName
	}
	if c.SchoolID == 0 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("missing school id of course %s", c.Name))
	}

	return nil
}

func toCourseEntity(currentCourses map[pgtype.Text]*entities_bob.Course, p *Course) (*entities_bob.Course, error) {
	var r *entities_bob.Course
	r, ok := currentCourses[database.Text(p.ID)]
	if !ok {
		r = &entities_bob.Course{}
		database.AllNullEntity(r)
	}

	var (
		grade int
		err   error
	)
	if p.Grade != "" {
		grade, err = i18n.ConvertStringGradeToInt(p.Country, p.Grade)
		if err != nil {
			return nil, err
		}
	}

	err = multierr.Combine(
		r.ID.Set(p.ID),
		r.Name.Set(p.Name),
		r.Country.Set(p.Country.String()),
		r.Subject.Set(p.Subject.String()),
		r.Grade.Set(grade),
		r.DisplayOrder.Set(p.DisplayOrder),
		r.Icon.Set(p.Icon),
		r.SchoolID.Set(p.SchoolID),
		r.CourseType.Set(pb_bob.COURSE_TYPE_CONTENT.String()),
		r.DeletedAt.Set(nil), // make sure this obj wil update successfully
		r.Status.Set(cpb.CourseStatus_COURSE_STATUS_NONE.String()),
	)

	if p.Status != cpb.CourseStatus_COURSE_STATUS_NONE {
		err = multierr.Append(err, r.Status.Set(p.Status.String()))
	}

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *CourseService) UpsertCourseClasses(ctx context.Context, req *pb.CoursesUpdated) (*pb.UpsertCourseClassesResponse, error) {
	if err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		var schoolIDs []int
		for _, course := range req.GetReq().GetCourses() {
			if !inArrayInt(int(course.SchoolId), schoolIDs) {
				schoolIDs = append(schoolIDs, int(course.SchoolId))
			}
		}
		if len(schoolIDs) == 0 {
			return nil
		}
		schoolClasses, err := s.ClassRepo.RetrieveClassIDsBySchoolIDs(ctx, s.DBTrace, schoolIDs)
		if err != nil {
			return errors.Wrapf(err, "c.ClassRepo.RetrieveClassIDsBySchoolID: schoolIDs: %v", schoolIDs)
		}
		if len(schoolClasses) == 0 {
			return nil
		}

		var cc []*entities_bob.CourseClass
		for _, course := range req.GetReq().GetCourses() {
			classes := schoolClasses[int(course.SchoolId)]
			if len(classes) == 0 {
				continue
			}
			for _, classID := range classes {
				c := &entities_bob.CourseClass{}
				database.AllNullEntity(c)
				err := multierr.Combine(
					c.ClassID.Set(classID),
					c.CourseID.Set(course.Id),
				)
				if err != nil {
					return fmt.Errorf("multierr.Combine: %w", err)
				}
				cc = append(cc, c)
			}
		}
		if err := s.CourseClassRepo.UpsertV2(ctx, s.DBTrace, cc); err != nil {
			return errors.Wrap(err, "c.CourseClassRepo.Upsert")
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("UpsertCourseClasses: %w", err)
	}
	return &pb.UpsertCourseClassesResponse{Successful: true}, nil
}

func (s *CourseService) UpsertQuizSets(ctx context.Context, req *pb.UpsertQuizSetsRequest) (*pb.UpsertQuizSetsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "func deleted")
}

// deleted
func (s *CourseService) UpsertQuestions(ctx context.Context, req *pb.UpsertQuestionsRequest) (*pb.UpsertQuestionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "func deleted")
}

func (s *CourseService) CreateBrightCoveUploadUrl(ctx context.Context, req *pb.CreateBrightCoveUploadUrlRequest) (*pb.CreateBrightCoveUploadUrlResponse, error) {
	resp, err := s.BrightcoveService.CreateBrightCoveUploadUrl(ctx, &ypb.CreateBrightCoveUploadUrlRequest{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateBrightCoveUploadUrlResponse{
		SignedUrl:     resp.SignedUrl,
		ApiRequestUrl: resp.ApiRequestUrl,
		VideoId:       resp.VideoId,
	}, err
}

func (s *CourseService) FinishUploadBrightCove(ctx context.Context, req *pb.FinishUploadBrightCoveRequest) (*pb.FinishUploadBrightCoveResponse, error) {
	_, err := s.BrightcoveService.FinishUploadBrightCove(ctx, &ypb.FinishUploadBrightCoveRequest{
		ApiRequestUrl: req.ApiRequestUrl,
		VideoId:       req.VideoId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.FinishUploadBrightCoveResponse{}, nil
}

func (s *CourseService) SyncCourse(ctx context.Context, req []*npb.EventMasterRegistration_Course) error {
	upsertCourses := []*Course{}
	deleteCourses := []string{}
	for _, r := range req {
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			upsertCourses = append(upsertCourses, &Course{
				ID:           r.CourseId,
				Name:         r.CourseName,
				Country:      pb_bob.COUNTRY_JP,
				Subject:      pb_bob.SUBJECT_ENGLISH,
				SchoolID:     constants.JPREPSchool,
				Grade:        i18n.OutGradeMap[pb_bob.COUNTRY_JP][0],
				DisplayOrder: 1,
				Status:       r.Status,
			})
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteCourses = append(deleteCourses, r.CourseId)
		}
	}

	var err error
	if len(upsertCourses) != 0 {
		errU := s.upsertCoursesV2(ctx, &UpsertCoursesV2Opt{
			Courses: upsertCourses,
		})
		if errU != nil {
			err = multierr.Append(err, fmt.Errorf("s.UpsertCourses: %w", errU))
		}
	}

	if len(deleteCourses) != 0 {
		_, errD := s.DeleteCourses(ctx, &pb.DeleteCoursesRequest{
			CourseIds: deleteCourses,
		})
		if errD != nil {
			err = multierr.Append(err, fmt.Errorf("s.DeleteCourses: %w", errD))
		}
	}

	return err
}

func (s *CourseService) CourseIDsByClass(ctx context.Context, classIDs []int32) (mapByClass map[int32][]string, err error) {
	result, err := s.CourseClassRepo.FindByClassIDs(ctx, s.DBTrace, database.Int4Array(classIDs))
	if err != nil {
		return nil, fmt.Errorf("err s.CourseClassRepo.FindByClassIDs: %w", err)
	}

	mapByClass = make(map[int32][]string)
	for classID, courseIDs := range result {
		var ids []string
		_ = courseIDs.AssignTo(&ids)
		mapByClass[classID.Int] = ids
	}
	return mapByClass, nil
}

func (s *CourseService) SyncAcademicYear(ctx context.Context, req []*npb.EventMasterRegistration_AcademicYear) error {
	return database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		for _, r := range req {
			switch r.ActionKind {
			case npb.ActionKind_ACTION_KIND_UPSERTED:
				academicYear := &entities_bob.AcademicYear{}
				database.AllNullEntity(academicYear)

				err := multierr.Combine(
					academicYear.ID.Set(r.AcademicYearId),
					academicYear.SchoolID.Set(constants.JPREPSchool),
					academicYear.Name.Set(r.Name),
					academicYear.StartYearDate.Set(r.StartYearDate.AsTime()),
					academicYear.EndYearDate.Set(r.EndYearDate.AsTime()),
					academicYear.Status.Set(entities_bob.AcademicYearStatusActive),
				)
				if err != nil {
					return fmt.Errorf("err set academicYear %s: %w", r.AcademicYearId, err)
				}

				err = s.AcademicYearRepo.Create(ctx, tx, academicYear)
				if err != nil {
					return fmt.Errorf("err AcademicYearRepo.Create %s: %w", r.AcademicYearId, err)
				}
			case npb.ActionKind_ACTION_KIND_DELETED:
				// noop
			}
		}

		return nil
	})
}

func (s *CourseService) UpdateAcademicYear(ctx context.Context, opts []*repositories.UpdateAcademicYearOpts) error {
	return database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		return s.CourseRepo.UpdateAcademicYear(ctx, tx, opts)
	})
}
