package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// NewCourseReaderService returns new *CourseReaderService
func NewCourseReaderService(
	eurekaDBTrace *database.DBTrace,
	db database.Ext, studyPlanReaderSvc pb.StudyPlanReaderServiceClient,
	flashcardReaderSvc pb.FlashCardReaderServiceClient,
	eurekaStudyPlanReader pb.StudyPlanReaderServiceClient,
	env string,
) *CourseReaderService {
	return &CourseReaderService{
		EurekaDBTrace:          eurekaDBTrace,
		DB:                     db,
		FlashCardReader:        flashcardReaderSvc,
		Env:                    env,
		BookRepo:               &repositories.BookRepo{},
		LessonRepo:             &repositories.LessonRepo{},
		LessonGroupRepo:        &repositories.LessonGroupRepo{},
		MediaRepo:              &repositories.MediaRepo{},
		CourseRepo:             &repositories.CourseRepo{},
		TeacherRepo:            &repositories.TeacherRepo{},
		QuizSetRepo:            &repositories.QuizSetRepo{},
		QuizRepo:               &repositories.QuizRepo{},
		StudyPlanReaderService: studyPlanReaderSvc,
		TopicRepo:              &repositories.TopicRepo{},
		ShuffledQuizSetRepo:    &repositories.ShuffledQuizSetRepo{},
		EurekaStudyPlanReader:  eurekaStudyPlanReader,
	}
}

// CourseReaderService will handle the api endpoint which is in resposibility for read data
type CourseReaderService struct {
	EurekaDBTrace database.Ext
	bpb.UnimplementedCourseReaderServiceServer
	DB                  database.Ext
	EurekaChapterReader pb.ChapterReaderServiceClient
	FlashCardReader     pb.FlashCardReaderServiceClient
	Env                 string
	BookRepo            interface {
		RetrieveBookTreeByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*repositories.BookTreeInfo, error)
	}
	LessonRepo interface {
		Find(context.Context, database.QueryExecer, *repositories.LessonFilter) ([]*entities.Lesson, error)
	}
	LessonGroupRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text) (*entities.LessonGroup, error)
		GetMedias(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text, pgtype.Int4, pgtype.Text) (entities.Medias, error)
	}
	MediaRepo interface {
		RetrieveByIDs(context.Context, database.QueryExecer, pgtype.TextArray) ([]*entities.Media, error)
	}
	CourseRepo interface {
		RetrieveCourses(ctx context.Context, db database.QueryExecer, q *repositories.CourseQuery) (entities.Courses, error)
	}
	TeacherRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}
	QuizSetRepo interface {
		GetQuizExternalIDs(ctx context.Context, db database.QueryExecer, loID pgtype.Text, limit pgtype.Int8, offset pgtype.Int8) ([]string, error)
	}
	QuizRepo interface {
		GetByExternalIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, loID pgtype.Text) (entities.Quizzes, error)
	}

	ShuffledQuizSetRepo interface {
		CalculateHigestSubmissionScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*repositories.CalculateHighestScoreResponse, error)
	}

	StudyPlanReaderService interface {
		GetBookIDsBelongsToStudentStudyPlan(context.Context, *pb.GetBookIDsBelongsToStudentStudyPlanRequest, ...grpc.CallOption) (*pb.GetBookIDsBelongsToStudentStudyPlanResponse, error)
	}

	TopicRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
	}

	EurekaStudyPlanReader interface {
		GetLOHighestScoresByStudyPlanItemIDs(ctx context.Context, in *pb.GetLOHighestScoresByStudyPlanItemIDsRequest, opts ...grpc.CallOption) (*pb.GetLOHighestScoresByStudyPlanItemIDsResponse, error)
	}
}

// ListLessonMedias return list of medias
func (c *CourseReaderService) ListLessonMedias(ctx context.Context, req *bpb.ListLessonMediasRequest) (*bpb.ListLessonMediasResponse, error) {
	filter := &repositories.LessonFilter{}
	err := multierr.Combine(
		filter.LessonID.Set([]string{req.LessonId}),
		filter.TeacherID.Set(nil),
		filter.CourseID.Set(nil),
	)
	if err != nil {
		return nil, fmt.Errorf("ListLessonMedias.SetFilter: %v", err)
	}
	lesson, err := c.LessonRepo.Find(ctx, c.DB, filter)
	if err != nil {
		return nil, fmt.Errorf("ListLessonMedias.LessonRepo.Find: %v", err)
	}

	if len(lesson) == 0 {
		return &bpb.ListLessonMediasResponse{}, nil
	}

	limit := database.Int4(int32(req.Paging.Limit))
	offset := pgtype.Text{}
	if len(req.Paging.GetOffsetString()) == 0 {
		offset.Set(nil)
	} else {
		offset.Set(req.Paging.GetOffsetString())
	}

	medias, err := c.LessonGroupRepo.GetMedias(ctx, c.DB, lesson[0].LessonGroupID, lesson[0].CourseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListLessonMedias.LessonGroupRepo.GetMedias: %v", err)
	}

	if len(medias) == 0 {
		return &bpb.ListLessonMediasResponse{}, nil
	}

	mediasPb := []*bpb.Media{}
	for _, media := range medias {
		mediaPb, err := toMediaPbV1(media)
		if err != nil {
			return nil, err
		}
		mediasPb = append(mediasPb, mediaPb)
	}

	resp := &bpb.ListLessonMediasResponse{
		Items: mediasPb,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: medias[len(medias)-1].MediaID.String,
			},
		},
	}

	return resp, nil
}

func toCommentsV1(src pgtype.JSONB) ([]*bpb.Comment, error) {
	var comments []entities.Comment
	err := src.AssignTo(&comments)
	if err != nil {
		return nil, err
	}
	dst := make([]*bpb.Comment, 0, len(comments))
	for _, comment := range comments {
		dst = append(dst, &bpb.Comment{
			Comment: comment.Comment,
		})
	}
	return dst, nil
}

func toMediaPbV1(src *entities.Media) (*bpb.Media, error) {
	comments, err := toCommentsV1(src.Comments)
	if err != nil {
		return nil, err
	}

	var convertedImages []*entities.ConvertedImage
	if err := src.ConvertedImages.AssignTo(&convertedImages); err != nil {
		return nil, err
	}
	pbImages := make([]*bpb.ConvertedImage, 0, len(convertedImages))
	for _, c := range convertedImages {
		pbImages = append(pbImages, &bpb.ConvertedImage{
			Width:    c.Width,
			Height:   c.Height,
			ImageUrl: c.ImageURL,
		})
	}

	return &bpb.Media{
		MediaId:   src.MediaID.String,
		Name:      src.Name.String,
		Resource:  src.Resource.String,
		CreatedAt: timestamppb.New(src.CreatedAt.Time),
		UpdatedAt: timestamppb.New(src.UpdatedAt.Time),
		Comments:  comments,
		Type:      bpb.MediaType(bpb.MediaType_value[src.Type.String]),
		Images:    pbImages,
	}, nil
}

func (c *CourseReaderService) ListCourses(ctx context.Context, req *bpb.ListCoursesRequest) (*bpb.ListCoursesResponse, error) {
	args := &repositories.CourseQuery{
		Limit:  10,
		Offset: 0,
	}

	if req.Filter != nil {
		if len(req.Filter.Ids) > 0 {
			args.IDs = req.Filter.Ids
		}

		if req.Filter.Country != cpb.Country_COUNTRY_NONE {
			args.Countries = []string{req.Filter.Country.String()}
		}

		if req.Filter.Subject != cpb.Subject_SUBJECT_NONE {
			args.Subject = req.Filter.Subject.String()
		}

		if req.Filter.Grade != 0 {
			args.Grade = int(req.Filter.Grade)
		}

		if req.Filter.SchoolId != 0 {
			args.SchoolIDs = []int{int(req.Filter.SchoolId)}
		}
	}

	if len(req.GetKeyword()) > 0 {
		args.Keyword = req.GetKeyword()
	}

	switch interceptors.UserGroupFromContext(ctx) {
	case entities.UserGroupTeacher:
		if len(args.SchoolIDs) == 0 {
			teacher, err := c.TeacherRepo.FindByID(ctx, c.DB, database.Text(interceptors.UserIDFromContext(ctx)))
			if err != nil {
				return nil, err
			}

			_ = teacher.SchoolIDs.AssignTo(&args.SchoolIDs)
		}
	}

	if paging := req.Paging; paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = int(limit)
		}
		if c := paging.GetOffsetCombined(); c != nil {
			if c.OffsetInteger != 0 {
				args.Offset = int(c.OffsetInteger)
			}
		}
	}

	courses, err := c.CourseRepo.RetrieveCourses(ctx, c.DB, args)
	if err != nil {
		return nil, fmt.Errorf("err c.CourseRepo.RetrieveCourses: %w", err)
	}
	if len(courses) == 0 {
		return &bpb.ListCoursesResponse{}, nil
	}

	pbCourses := make([]*cpb.Course, 0, len(courses))
	for _, course := range courses {
		pbCourses = append(pbCourses, &cpb.Course{
			Info: &cpb.ContentBasicInfo{
				Id:           course.ID.String,
				Name:         course.Name.String,
				Country:      cpb.Country(cpb.Country_value[course.Country.String]),
				Subject:      cpb.Subject(cpb.Subject_value[course.Subject.String]),
				Grade:        int32(course.Grade.Int),
				SchoolId:     course.SchoolID.Int,
				DisplayOrder: int32(course.DisplayOrder.Int),
				UpdatedAt:    timestamppb.New(course.UpdatedAt.Time),
				CreatedAt:    timestamppb.New(course.CreatedAt.Time),
				IconUrl:      course.Icon.String,
			},
			CourseStatus: cpb.CourseStatus_COURSE_STATUS_ACTIVE,
		})
	}

	return &bpb.ListCoursesResponse{
		NextPage: &cpb.Paging{
			Limit: uint32(args.Limit),
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetInteger: int64(args.Offset + args.Limit),
				},
			},
		},
		Items: pbCourses,
	}, nil
}

func (c *CourseReaderService) ListCoursesByLocations(ctx context.Context, req *bpb.ListCoursesByLocationsRequest) (*bpb.ListCoursesByLocationsResponse, error) {
	args := &repositories.CourseQuery{
		Limit:  10,
		Offset: 0,
	}

	if req.Filter != nil {
		if len(req.Filter.Ids) > 0 {
			args.IDs = req.Filter.Ids
		}

		if req.Filter.Country != cpb.Country_COUNTRY_NONE {
			args.Countries = []string{req.Filter.Country.String()}
		}

		if req.Filter.Subject != cpb.Subject_SUBJECT_NONE {
			args.Subject = req.Filter.Subject.String()
		}

		if req.Filter.Grade != 0 {
			args.Grade = int(req.Filter.Grade)
		}

		if req.Filter.SchoolId != 0 {
			args.SchoolIDs = []int{int(req.Filter.SchoolId)}
		}
	}

	if len(req.LocationIds) > 0 {
		args.LocationIDs = req.LocationIds
	}

	if len(req.GetKeyword()) > 0 {
		args.Keyword = req.GetKeyword()
	}

	switch interceptors.UserGroupFromContext(ctx) {
	case entities.UserGroupTeacher:
		if len(args.SchoolIDs) == 0 {
			teacher, err := c.TeacherRepo.FindByID(ctx, c.DB, database.Text(interceptors.UserIDFromContext(ctx)))
			if err != nil {
				return nil, err
			}

			_ = teacher.SchoolIDs.AssignTo(&args.SchoolIDs)
		}
	}

	if paging := req.Paging; paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = int(limit)
		}
		if c := paging.GetOffsetCombined(); c != nil {
			if c.OffsetInteger != 0 {
				args.Offset = int(c.OffsetInteger)
			}
		}
	}

	courses, err := c.CourseRepo.RetrieveCourses(ctx, c.DB, args)
	if err != nil {
		return nil, fmt.Errorf("err c.CourseRepo.RetrieveCourses: %w", err)
	}
	if len(courses) == 0 {
		return &bpb.ListCoursesByLocationsResponse{}, nil
	}

	pbCourses := make([]*cpb.Course, 0, len(courses))
	for _, course := range courses {
		pbCourses = append(pbCourses, &cpb.Course{
			Info: &cpb.ContentBasicInfo{
				Id:           course.ID.String,
				Name:         course.Name.String,
				Country:      cpb.Country(cpb.Country_value[course.Country.String]),
				Subject:      cpb.Subject(cpb.Subject_value[course.Subject.String]),
				Grade:        int32(course.Grade.Int),
				SchoolId:     course.SchoolID.Int,
				DisplayOrder: int32(course.DisplayOrder.Int),
				UpdatedAt:    timestamppb.New(course.UpdatedAt.Time),
				CreatedAt:    timestamppb.New(course.CreatedAt.Time),
				IconUrl:      course.Icon.String,
			},
			CourseStatus: cpb.CourseStatus_COURSE_STATUS_ACTIVE,
		})
	}

	return &bpb.ListCoursesByLocationsResponse{
		NextPage: &cpb.Paging{
			Limit: uint32(args.Limit),
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetInteger: int64(args.Offset + args.Limit),
				},
			},
		},
		Items: pbCourses,
	}, nil
}

func (c *CourseReaderService) validateRetrieveFlashCardStudyProgressRequest(req *bpb.RetrieveFlashCardStudyProgressRequest) error {
	if req.StudySetId == "" {
		return status.Error(codes.InvalidArgument, "req must have study set id")
	}
	if req.StudentId == "" {
		return status.Error(codes.InvalidArgument, "req must have student id")
	}

	if req.Paging == nil {
		return status.Error(codes.InvalidArgument, "req must have paging field")
	}

	if req.Paging.GetOffsetInteger() <= 0 {
		return status.Error(codes.InvalidArgument, "offset must be positive")
	}

	if req.Paging.Limit <= 0 {
		req.Paging.Limit = 100
	}
	return nil
}

func (c *CourseReaderService) RetrieveFlashCardStudyProgress(ctx context.Context, req *bpb.RetrieveFlashCardStudyProgressRequest) (*bpb.RetrieveFlashCardStudyProgressResponse, error) {
	mdctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Errorf("CourseReaderService.RetrieveFlashCardStudyProgress.GetOutgoingContext: %w", err).Error())
	}

	resp, err := c.FlashCardReader.RetrieveFlashCardStudyProgress(mdctx, &pb.RetrieveFlashCardStudyProgressRequest{
		StudySetId: req.StudySetId,
		StudentId:  req.StudentId,
		Paging:     req.Paging,
	})
	if err != nil {
		return nil, err
	}

	items := []*bpb.FlashcardQuizzes{}
	for _, item := range resp.Items {
		items = append(items, &bpb.FlashcardQuizzes{
			Item:   item.Item,
			Status: bpb.FlashcardQuizStudyStatus(item.Status),
		})
	}

	return &bpb.RetrieveFlashCardStudyProgressResponse{
		NextPage:      resp.NextPage,
		StudySetId:    resp.StudySetId,
		Items:         items,
		StudyingIndex: resp.StudyingIndex,
	}, nil
}

func (c *CourseReaderService) RetrieveBookTreeByTopicIDs(ctx context.Context, req *bpb.RetrieveBookTreeByTopicIDsRequest) (*bpb.RetrieveBookTreeByTopicIDsResponse, error) {
	m := make(map[string]bool)
	topics, err := c.TopicRepo.RetrieveByIDs(ctx, c.EurekaDBTrace, database.TextArray(req.TopicIds))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve topic by ids: %w", err).Error())
	}
	for _, topic := range topics {
		m[topic.ID.String] = true
	}
	for _, topicID := range req.TopicIds {
		if _, ok := m[topicID]; !ok {
			return nil, status.Errorf(codes.NotFound, "topic %s not exist", topicID)
		}
	}

	infos, err := c.BookRepo.RetrieveBookTreeByTopicIDs(ctx, c.EurekaDBTrace, database.TextArray(req.TopicIds))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve los: %w", err).Error())
	}
	infosRes := make([]*bpb.RetrieveBookTreeByTopicIDsResponse_Info, 0, len(infos))
	for _, info := range infos {
		var id *wrapperspb.StringValue
		if info.LoID.Status == pgtype.Present {
			id = wrapperspb.String(info.LoID.String)
		} else {
			id = nil
		}
		infosRes = append(infosRes, &bpb.RetrieveBookTreeByTopicIDsResponse_Info{
			LoId:                id,
			TopicId:             info.TopicID.String,
			ChapterId:           info.ChapterID.String,
			LoDisplayOrder:      int32(info.LoDisplayOrder.Int),
			TopicDisplayOrder:   int32(info.TopicDisplayOrder.Int),
			ChapterDispalyOrder: int32(info.ChapterDisplayOrder.Int),
		})
	}
	return &bpb.RetrieveBookTreeByTopicIDsResponse{
		Infos: infosRes,
	}, nil
}

func ToTopicPbV1(p *entities.Topic) *cpb.Topic {
	topic := &cpb.Topic{
		Info: &cpb.ContentBasicInfo{
			Id:           p.ID.String,
			Name:         p.Name.String,
			Country:      cpb.Country(cpb.Country_value[p.Country.String]),
			Subject:      cpb.Subject(cpb.Subject_value[p.Subject.String]),
			Grade:        int32(p.Grade.Int),
			SchoolId:     p.SchoolID.Int,
			DisplayOrder: int32(p.DisplayOrder.Int),
			IconUrl:      p.IconURL.String,
			CreatedAt:    &timestamppb.Timestamp{Seconds: p.CreatedAt.Time.Unix()},
			UpdatedAt:    &timestamppb.Timestamp{Seconds: p.UpdatedAt.Time.Unix()},
		},
		Type:        cpb.TopicType(cpb.TopicType_value[p.TopicType.String]),
		Status:      cpb.TopicStatus(cpb.TopicStatus_value[p.Status.String]),
		ChapterId:   p.ChapterID.String,
		Instruction: p.Instruction.String,
	}

	numAttachment := len(p.AttachmentNames.Elements)
	if n := len(p.AttachmentURLs.Elements); n < numAttachment {
		numAttachment = n
	}

	for i := 0; i < numAttachment; i++ {
		topic.Attachments = append(topic.Attachments, &cpb.Attachment{
			Name: p.AttachmentNames.Elements[i].String,
			Url:  p.AttachmentURLs.Elements[i].String,
		})
	}

	return topic
}
