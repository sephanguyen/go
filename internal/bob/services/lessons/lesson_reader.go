package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/bob/services/log"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonReaderServices struct {
	bpb.UnimplementedLessonReaderServiceServer
	DB                         database.Ext
	Env                        string
	UnleashClientIns           unleashclient.ClientInstance
	VirtualClassRoomLogService *log.VirtualClassRoomLogService
	LessonRepo                 interface {
		GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, queryEnhancers ...repositories.QueryEnhancer) ([]string, error)
		Retrieve(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) ([]*entities.Lesson, uint32, string, uint32, error)
		FindPreviousPageOffset(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (string, error)
		CountLesson(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (int64, error)
		GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
		GetCourseIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
		// Course's find lesson endpoint
		FindLessonWithTime(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, scheduling_status pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		FindLessonWithTimeAndLocations(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, scheduling_status pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		FindLessonJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, scheduling_status pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
		FindLessonJoinedWithLocations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, scheduling_status pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesByUser(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) (entities.LessonMemberStates, error)
		GetLessonMemberStates(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.LessonMemberStates, error)
		CourseAccessible(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error)
	}
	SchoolAdminRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities.SchoolAdmin, error)
	}
	MediaRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error)
	}
	UserRepo interface {
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
	}
	LessonRoomStateRepo interface {
		GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*domain.LessonRoomState, error)
	}
	CourseClassRepo interface {
		Find(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapCourseIDsByClassID map[pgtype.Int4]pgtype.TextArray, err error)
	}
	ClassRepo interface {
		FindJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities.Class, error)
	}
}

func (s *LessonReaderServices) GetStreamingLearners(ctx context.Context, req *bpb.GetStreamingLearnersRequest) (*bpb.GetStreamingLearnersResponse, error) {
	if req.GetLessonId() == "" {
		return nil, fmt.Errorf("LessonID must not empty")
	}
	ids, err := s.LessonRepo.GetStreamingLearners(ctx, s.DB, database.Text(req.LessonId))
	if err != nil {
		return nil, fmt.Errorf("s.LessonStreamRepo.GetStreamingLearners: %w", err)
	}

	return &bpb.GetStreamingLearnersResponse{LearnerIds: ids}, nil
}

func (s *LessonReaderServices) RetrieveLessons(ctx context.Context, req *bpb.RetrieveLessonsRequest) (*bpb.RetrieveLessonsResponse, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.Internal, "missing paging info")
	}
	args := filterArgsFromRequest(req)
	if req.GetPaging().GetOffsetString() != "" {
		err := args.LessonID.Set(req.Paging.GetOffsetString())
		if err != nil {
			return nil, fmt.Errorf("cannot set ListLessonArgs.LessonID: %s", err)
		}
	}

	userID := interceptors.UserIDFromContext(ctx)
	group, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if group == constant.UserGroupSchoolAdmin {
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(userID))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		args.SchoolID = schoolAdmin.SchoolID
	}

	lessons, total, prePageID, preTotal, err := s.LessonRepo.Retrieve(ctx, s.DB, &args)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if preTotal <= req.Paging.Limit {
		prePageID = ""
	}

	items := []*bpb.RetrieveLessonsResponse_Lesson{}
	if len(lessons) == 0 {
		return &bpb.RetrieveLessonsResponse{}, nil
	}
	for _, l := range lessons {
		teacherIDs, err := s.LessonRepo.GetTeacherIDsOfLesson(ctx, s.DB, l.LessonID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		l.TeacherIDs.TeacherIDs = teacherIDs
		courseIDs, err := s.LessonRepo.GetCourseIDsOfLesson(ctx, s.DB, l.LessonID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		l.CourseIDs.CourseIDs = courseIDs

		items = append(items, toLessonPb(l))
	}

	lastItem := lessons[len(lessons)-1]

	return &bpb.RetrieveLessonsResponse{
		Items: items,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem.LessonID.String,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: prePageID,
			},
		},
		TotalLesson: total,
	}, nil
}

func filterArgsFromRequest(req *bpb.RetrieveLessonsRequest) repositories.ListLessonArgs {
	args := repositories.ListLessonArgs{
		Limit:            req.Paging.Limit,
		LessonID:         pgtype.Text{Status: pgtype.Null},
		SchoolID:         pgtype.Int4{Status: pgtype.Null},
		Courses:          pgtype.TextArray{Status: pgtype.Null},
		StartTime:        pgtype.Timestamptz{Status: pgtype.Null},
		EndTime:          pgtype.Timestamptz{Status: pgtype.Null},
		StatusNotStarted: pgtype.Text{Status: pgtype.Null},
		StatusInProcess:  pgtype.Text{Status: pgtype.Null},
		StatusCompleted:  pgtype.Text{Status: pgtype.Null},
		KeyWord:          pgtype.Text{Status: pgtype.Null},
	}

	if courses := req.GetFilter().GetCourseIds(); len(courses) > 0 {
		args.Courses = database.TextArray(courses)
	}

	if startTime := req.GetFilter().GetStartTime(); startTime != nil {
		args.StartTime = database.Timestamptz(startTime.AsTime())
	}

	if endTime := req.GetFilter().GetEndTime(); endTime != nil {
		args.EndTime = database.Timestamptz(endTime.AsTime())
	}

	if status := req.GetFilter().GetLessonStatus(); len(status) > 0 {
		for _, state := range status {
			switch state {
			case cpb.LessonStatus_LESSON_STATUS_COMPLETED:
				args.StatusCompleted = database.Text("COMPLETED")
			case cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS:
				args.StatusInProcess = database.Text("IN_PROGRESS")
			case cpb.LessonStatus_LESSON_STATUS_NOT_STARTED:
				args.StatusNotStarted = database.Text("NOT_STARTED")
			}
		}
	}

	if keyWord := req.GetKeyword(); keyWord != "" {
		args.KeyWord = database.Text(keyWord)
	}

	return args
}

func toLessonPb(l *entities.Lesson) *bpb.RetrieveLessonsResponse_Lesson {
	return &bpb.RetrieveLessonsResponse_Lesson{
		Id:         l.LessonID.String,
		Name:       l.Name.String,
		StartTime:  timestamppb.New(l.StartTime.Time),
		EndTime:    timestamppb.New(l.EndTime.Time),
		CourseIds:  database.FromTextArray(l.CourseIDs.CourseIDs),
		TeacherIds: database.FromTextArray(l.TeacherIDs.TeacherIDs),
	}
}

func (s *LessonReaderServices) GetLiveLessonState(ctx context.Context, req *bpb.LiveLessonStateRequest) (*bpb.LiveLessonStateResponse, error) {
	if len(strings.TrimSpace(req.Id)) == 0 {
		return nil, status.Error(codes.Internal, fmt.Sprintf("LessonID can't empty"))
	}
	// get lesson
	lesson, err := s.getLiveLesson(ctx, database.Text(req.Id), includeLearnerIDs(), includeTeacherIDs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// check permission for each user type
	userID := database.Text(interceptors.UserIDFromContext(ctx))
	var learnerStates entities.LessonMemberStates
	if !lesson.TeacherIDs.HaveID(userID) && !lesson.LearnerIDs.HaveID(userID) {
		// user group
		userGroup, err := s.UserRepo.UserGroup(ctx, s.DB, userID)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.UserGroup: %w", err)
		}

		if userGroup == entities.UserGroupStudent {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("user could not get live lesson state: %s", err))
		}
	}

	learnerStates, err = s.LessonMemberRepo.GetLessonMemberStates(ctx, s.DB, lesson.LessonID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("LessonMemberRepo.GetLessonMemberStates: %s", err))
	}

	// create live lesson state by room's states and learner's states
	lls, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, learnerStates)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// get media of current material
	var media *entities.Media
	if lls.RoomState != nil && lls.RoomState.CurrentMaterial != nil {
		medias, err := s.MediaRepo.RetrieveByIDs(ctx, s.DB, database.TextArray([]string{lls.RoomState.CurrentMaterial.MediaID}))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("MediaRepo.RetrieveByIDs: %s", err))
		}
		if len(medias) != 0 {
			media = medias[0]
		}
	}

	// create response
	res, err := liveLessonStateResponseFromLiveLessonState(lls, media)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res.CurrentTime = timestamppb.New(time.Now())

	// log for virtual classroom
	if err = s.VirtualClassRoomLogService.LogWhenGetRoomState(ctx, database.Text(req.Id)); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenGetRoomState: could not log this activity",
			zap.String("lesson_id", req.Id),
			zap.String("user_ID", userID.String),
			zap.Error(err),
		)
	}

	if err = s.liveLessonStateResponseFromGetLessonRoomState(ctx, res, lesson.LessonID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (s *LessonReaderServices) liveLessonStateResponseFromGetLessonRoomState(ctx context.Context, res *bpb.LiveLessonStateResponse, lessonID pgtype.Text) error {
	spotlight := &bpb.LiveLessonState_Spotlight{}
	lessonRoomState, err := s.LessonRoomStateRepo.GetLessonRoomStateByLessonID(ctx, s.DB, lessonID)
	if err == domain.ErrNotFound {
		res.Spotlight = spotlight
		res.WhiteboardZoomState = toWhiteboardZoomStateBp(new(virDomain.WhiteboardZoomState).SetDefault())
		res.Recording = &bpb.LiveLessonState_Recording{}
		res.CurrentMaterial = &bpb.LiveLessonState_CurrentMaterial{}
		return nil
	} else if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error in LessonRoomStateRepo.GetLessonRoomStateByLessonID, lesson %s: %s", lessonID.String, err))
	}
	if lessonRoomState.SpotlightedUser != "" {
		spotlight.IsSpotlight = true
		spotlight.UserId = lessonRoomState.SpotlightedUser
	}
	res.Spotlight = spotlight
	res.WhiteboardZoomState = toWhiteboardZoomStateBp(lessonRoomState.WhiteboardZoomState)
	if lessonRoomState.Recording != nil {
		res.Recording = &bpb.LiveLessonState_Recording{
			IsRecording: lessonRoomState.Recording.IsRecording,
			Creator:     lessonRoomState.Recording.Creator,
		}
	}
	if lessonRoomState.CurrentPolling != nil {
		polling := lessonRoomState.CurrentPolling
		options := make([]*bpb.LiveLessonState_PollingOption, 0, len(polling.Options))
		for _, option := range polling.Options {
			options = append(options, &bpb.LiveLessonState_PollingOption{
				Answer:    option.Answer,
				IsCorrect: option.IsCorrect,
				Content:   option.Content,
			})
		}

		CurrentPolling := &bpb.LiveLessonState_CurrentPolling{
			Options:   options,
			Status:    bpb.PollingState(bpb.PollingState_value[string(polling.Status)]),
			CreatedAt: timestamppb.New(polling.CreatedAt),
			IsShared:  lessonRoomState.CurrentPolling.IsShared,
			Question:  lessonRoomState.CurrentPolling.Question,
		}
		if polling.StoppedAt != nil {
			CurrentPolling.StoppedAt = timestamppb.New(*polling.StoppedAt)
		}
		res.CurrentPolling = CurrentPolling
	}

	if lessonRoomState.CurrentMaterial != nil {
		material := lessonRoomState.CurrentMaterial
		currentMaterial := &bpb.LiveLessonState_CurrentMaterial{
			MediaId:   material.MediaID,
			UpdatedAt: timestamppb.New(material.UpdatedAt),
		}

		medias, err := s.MediaRepo.RetrieveByIDs(ctx, s.DB, database.TextArray([]string{material.MediaID}))
		if err != nil {
			return fmt.Errorf("error in MediaRepo.RetrieveByIDs, lesson %s: %w", lessonID.String, err)
		}
		if len(medias) != 0 {
			mediaMess, err := toMediaPb(medias[0])
			if err != nil {
				return err
			}

			currentMaterial.Data = mediaMess
		} else {
			return fmt.Errorf("media %s is null, lesson %s", material.MediaID, lessonID.String)
		}

		if material.VideoState != nil {
			currentMaterial.State = &bpb.LiveLessonState_CurrentMaterial_VideoState_{
				VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
					CurrentTime: durationpb.New(material.VideoState.CurrentTime.Duration()),
					PlayerState: bpb.PlayerState(bpb.PlayerState_value[string(material.VideoState.PlayerState)]),
				},
			}
		} else if material.AudioState != nil {
			currentMaterial.State = &bpb.LiveLessonState_CurrentMaterial_AudioState_{
				AudioState: &bpb.LiveLessonState_CurrentMaterial_AudioState{
					CurrentTime: durationpb.New(material.AudioState.CurrentTime.Duration()),
					PlayerState: bpb.PlayerState(bpb.PlayerState_value[string(material.AudioState.PlayerState)]),
				},
			}
		}

		res.CurrentMaterial = currentMaterial
	}

	return nil
}

func toWhiteboardZoomStateBp(w *virDomain.WhiteboardZoomState) *bpb.LiveLessonState_WhiteboardZoomState {
	res := new(bpb.LiveLessonState_WhiteboardZoomState)
	res.PdfScaleRatio = w.PdfScaleRatio
	res.CenterX = w.CenterX
	res.CenterY = w.CenterY
	res.PdfWidth = w.PdfWidth
	res.PdfHeight = w.PdfHeight
	return res
}

type getLiveLessonOption func(context.Context, *LessonReaderServices, *entities.Lesson) error

func includeLearnerIDs() getLiveLessonOption {
	return func(ctx context.Context, s *LessonReaderServices, lesson *entities.Lesson) error {
		learnerIDs, err := s.LessonRepo.GetLearnerIDsOfLesson(ctx, s.DB, lesson.LessonID)
		if err != nil {
			return fmt.Errorf("LessonRepo.GetLearnerIDsOfLesson: %s", err)
		}
		lesson.LearnerIDs.LearnerIDs = learnerIDs

		return nil
	}
}

func includeTeacherIDs() getLiveLessonOption {
	return func(ctx context.Context, s *LessonReaderServices, lesson *entities.Lesson) error {
		teacherIDs, err := s.LessonRepo.GetTeacherIDsOfLesson(ctx, s.DB, lesson.LessonID)
		if err != nil {
			return fmt.Errorf("LessonRepo.GetTeacherIDsOfLesson: %s", err)
		}
		lesson.TeacherIDs.TeacherIDs = teacherIDs

		return nil
	}
}

func (s *LessonReaderServices) getLiveLesson(ctx context.Context, lessonID pgtype.Text, opts ...getLiveLessonOption) (*entities.Lesson, error) {
	// lesson and list its teachers, learners
	lesson, err := s.LessonRepo.FindByID(ctx, s.DB, lessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.FindByID %s: %s", lessonID.String, err)
	}

	for _, opt := range opts {
		if err = opt(ctx, s, lesson); err != nil {
			return nil, err
		}
	}

	return lesson, nil
}

func toComments(src pgtype.JSONB) ([]*bpb.Comment, error) {
	if src.Status != pgtype.Present {
		return nil, nil
	}

	var comments []entities.Comment
	err := src.AssignTo(&comments)
	if err != nil {
		return nil, err
	}
	dst := make([]*bpb.Comment, 0, len(comments))
	for _, comment := range comments {
		dst = append(dst, &bpb.Comment{
			Comment:  comment.Comment,
			Duration: durationpb.New(time.Duration(comment.Duration) * time.Second),
		})
	}
	return dst, nil
}

func toImages(src pgtype.JSONB) ([]*bpb.ConvertedImage, error) {
	if src.Status != pgtype.Present {
		return nil, nil
	}

	var convertedImages []*entities.ConvertedImage
	if err := src.AssignTo(&convertedImages); err != nil {
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

	return pbImages, nil
}

func toMediaPb(src *entities.Media) (*bpb.Media, error) {
	comments, err := toComments(src.Comments)
	if err != nil {
		return nil, err
	}

	pbImages, err := toImages(src.ConvertedImages)
	if err != nil {
		return nil, err
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

func liveLessonStateResponseFromLiveLessonState(lls *LiveLessonState, media *entities.Media) (*bpb.LiveLessonStateResponse, error) {
	res := bpb.LiveLessonStateResponse{
		Id: lls.LessonID,
	}

	if lls.RoomState != nil {
		if lls.RoomState.CurrentMaterial != nil {
			material := lls.RoomState.CurrentMaterial
			currentMaterial := &bpb.LiveLessonState_CurrentMaterial{
				MediaId:   material.MediaID,
				UpdatedAt: timestamppb.New(material.UpdatedAt),
			}

			if media == nil {
				return nil, fmt.Errorf("media %s is null", material.MediaID)
			}
			mediaMess, err := toMediaPb(media)
			if err != nil {
				return nil, err
			}
			currentMaterial.Data = mediaMess
			if material.VideoState != nil {
				currentMaterial.State = &bpb.LiveLessonState_CurrentMaterial_VideoState_{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(material.VideoState.CurrentTime.Duration()),
						PlayerState: bpb.PlayerState(bpb.PlayerState_value[string(material.VideoState.PlayerState)]),
					},
				}
			} else if material.AudioState != nil {
				currentMaterial.State = &bpb.LiveLessonState_CurrentMaterial_AudioState_{
					AudioState: &bpb.LiveLessonState_CurrentMaterial_AudioState{
						CurrentTime: durationpb.New(material.AudioState.CurrentTime.Duration()),
						PlayerState: bpb.PlayerState(bpb.PlayerState_value[string(material.AudioState.PlayerState)]),
					},
				}
			}

			res.CurrentMaterial = currentMaterial
		}
		if lls.RoomState.CurrentPolling != nil {
			polling := lls.RoomState.CurrentPolling
			options := make([]*bpb.LiveLessonState_PollingOption, 0, len(polling.Options))
			for _, option := range polling.Options {
				options = append(options, &bpb.LiveLessonState_PollingOption{
					Answer:    option.Answer,
					IsCorrect: option.IsCorrect,
				})
			}

			CurrentPolling := &bpb.LiveLessonState_CurrentPolling{
				Options:   options,
				Status:    bpb.PollingState(bpb.PollingState_value[string(polling.Status)]),
				CreatedAt: timestamppb.New(polling.CreatedAt),
			}
			if !polling.StoppedAt.IsZero() {
				CurrentPolling.StoppedAt = timestamppb.New(polling.StoppedAt)
			}

			res.CurrentPolling = CurrentPolling
		}
		if lls.RoomState.Recording != nil {
			res.Recording = &bpb.LiveLessonState_Recording{
				IsRecording: lls.RoomState.Recording.IsRecording,
			}
			if lls.RoomState.Recording.Creator != nil {
				res.Recording.Creator = *lls.RoomState.Recording.Creator
			}
		}
	}
	if lls.UserStates != nil {
		res.UsersState = &bpb.LiveLessonStateResponse_UsersState{}
		if len(lls.UserStates.LearnersState) != 0 {
			learnerState := make([]*bpb.LiveLessonStateResponse_UsersState_LearnerState, 0, len(lls.UserStates.LearnersState))
			for _, state := range lls.UserStates.LearnersState {
				learnerState = append(
					learnerState,
					&bpb.LiveLessonStateResponse_UsersState_LearnerState{
						UserId: state.UserID,
						HandsUp: &bpb.LiveLessonState_HandsUp{
							Value:     state.HandsUp.Value,
							UpdatedAt: timestamppb.New(state.HandsUp.UpdatedAt),
						},
						Annotation: &bpb.LiveLessonState_Annotation{
							Value:     state.Annotation.Value,
							UpdatedAt: timestamppb.New(state.Annotation.UpdatedAt),
						},
						PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
							StringArrayValue: state.PollingAnswer.StringArrayValue,
							UpdatedAt:        timestamppb.New(state.PollingAnswer.UpdatedAt),
						},
						Chat: &bpb.LiveLessonState_Chat{
							Value:     state.Chat.Value,
							UpdatedAt: timestamppb.New(state.Chat.UpdatedAt),
						},
					},
				)
			}
			res.UsersState.Learners = learnerState
		}
	}

	return &res, nil
}

func toBpbLesson(src *repositories.LessonWithTime, teacher *cpb.BasicProfile) *bpb.Lesson {
	status := cpb.LessonStatus_LESSON_STATUS_NOT_STARTED
	if src.Lesson.StartTime.Time.Unix() >= timeutil.Now().Unix() {
		status = cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS
	}
	if src.Lesson.EndTime.Time.Unix() < timeutil.Now().Unix() || (src.Lesson.EndAt.Status == pgtype.Present) {
		status = cpb.LessonStatus_LESSON_STATUS_COMPLETED
	}

	l := &bpb.Lesson{
		LessonId:                 src.Lesson.LessonID.String,
		CourseId:                 src.Lesson.CourseID.String,
		PresetStudyPlanWeeklyIds: "",
		Topic: &cpb.Topic{
			Attachments: []*cpb.Attachment{},
		},
		StartTime: &timestamppb.Timestamp{Seconds: src.Lesson.StartTime.Time.Unix()},
		EndTime:   &timestamppb.Timestamp{Seconds: src.Lesson.EndTime.Time.Unix()},
		Status:    status,
		ZoomLink:  src.Lesson.ZoomLink.String,
	}
	if teacher != nil {
		l.Teacher = []*cpb.BasicProfile{teacher}
	}

	return l
}

func (s *LessonReaderServices) findStudentValidCourse(ctx context.Context, userID string) ([]string, error) {
	var validCourseIDs []string
	classes, err := s.ClassRepo.FindJoined(ctx, s.DB, database.Text(userID))
	if err != nil {
		return nil, err
	}
	classIDs := make([]int32, 0, len(classes))
	for _, class := range classes {
		classIDs = append(classIDs, class.ID.Int)
	}
	courseMapByClass, err := s.CourseClassRepo.Find(ctx, s.DB, database.Int4Array(classIDs))
	if err != nil {
		return validCourseIDs, err
	}
	for _, courseIDByClass := range courseMapByClass {
		var courseIDs []string
		courseIDByClass.AssignTo(&courseIDs)
		validCourseIDs = append(validCourseIDs, courseIDs...)
	}

	courseIDs, err := s.LessonMemberRepo.CourseAccessible(ctx, s.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("err rcv.LessonMemberRepo.CourseAccessible: %w", err)
	}
	validCourseIDs = append(validCourseIDs, courseIDs...)

	return validCourseIDs, nil
}

func (s *LessonReaderServices) RetrieveLiveLessonByLocations(ctx context.Context, req *bpb.RetrieveLiveLessonByLocationsRequest) (*bpb.RetrieveLiveLessonByLocationsResponse, error) {
	validCourseIDs := database.TextArray(req.CourseIds)
	from := database.TimestamptzFromPb(req.From)
	to := database.TimestamptzFromPb(req.To)
	locationIDs := database.TextArray(req.LocationIds)
	var limit, page int32
	if req.Pagination != nil {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	userID := interceptors.UserIDFromContext(ctx)
	group, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(userID))
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	var (
		lessons []*repositories.LessonWithTime
		total   pgtype.Int8
	)

	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled("BACKEND_Lesson_HandleShowOnlyPublishStatusForEndpointListLessonForTeacherStudent", s.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}
	schedulingStatus := pgtype.Text{Status: pgtype.Null}
	if isUnleashToggled {
		if err := schedulingStatus.Set(entities.LessonSchedulingStatusPublished); err != nil {
			return nil, services.ToStatusError(err)
		}
	}
	if group == constant.UserGroupStudent {
		courseIDs, err := s.findStudentValidCourse(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get course for student error: %w", err)
		}
		courseIDs = append(courseIDs, req.CourseIds...)
		validCourseIDs = database.TextArray(courseIDs)
	}
	// check for location ids
	if len(req.LocationIds) > 0 {
		switch group {
		case constant.UserGroupStudent:
			lessons, total, err = s.LessonRepo.FindLessonJoinedWithLocations(ctx, s.DB, database.Text(userID), &validCourseIDs, &from, &to, &locationIDs, limit, page, schedulingStatus)
		default:
			lessons, total, err = s.LessonRepo.FindLessonWithTimeAndLocations(ctx, s.DB, &validCourseIDs, &from, &to, &locationIDs, limit, page, schedulingStatus)
		}
		if err != nil {
			return nil, services.ToStatusError(err)
		}
	} else {
		switch group {
		case constant.UserGroupStudent:
			lessons, total, err = s.LessonRepo.FindLessonJoined(ctx, s.DB, database.Text(userID), &validCourseIDs, &from, &to, limit, page, schedulingStatus)
		default:
			lessons, total, err = s.LessonRepo.FindLessonWithTime(ctx, s.DB, &validCourseIDs, &from, &to, limit, page, schedulingStatus)
		}

		if err != nil {
			return nil, services.ToStatusError(err)
		}
	}
	//TODO: create filter struct instead
	teacherIDs := make([]string, 0, len(lessons))
	for _, lesson := range lessons {
		teacherIDs = append(teacherIDs, lesson.Lesson.TeacherID.String)
	}

	teacherProfilesMap, err := s.getTeacherProfileCpb(ctx, teacherIDs)
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	bpbLessons := toBpbLessons(lessons, teacherProfilesMap)
	return &bpb.RetrieveLiveLessonByLocationsResponse{
		Lessons: bpbLessons,
		Total:   int32(total.Int),
	}, nil
}

func toBpbLessons(lessons []*repositories.LessonWithTime, teacherMap map[string]*cpb.BasicProfile) []*bpb.Lesson {
	pbLessons := make([]*bpb.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		teacher := teacherMap[lesson.Lesson.TeacherID.String]

		bpbLesson := toBpbLesson(lesson, teacher)
		pbLessons = append(pbLessons, bpbLesson)
	}

	return pbLessons
}

func (s *LessonReaderServices) getTeacherProfileCpb(ctx context.Context, teacherIDs []string) (map[string]*cpb.BasicProfile, error) {
	if len(teacherIDs) == 0 {
		return nil, nil
	}

	teacherProfilesMap := make(map[string]*cpb.BasicProfile, len(teacherIDs))
	teachers, err := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray(teacherIDs))
	if err != nil {
		return nil, fmt.Errorf("UserRepo.Retrieve: %w", err)
	}

	for _, teacher := range teachers {
		basicProfile := toBasicProfileCpb(teacher)
		teacherProfilesMap[teacher.ID.String] = basicProfile
	}

	return teacherProfilesMap, nil
}

func toBasicProfileCpb(src *entities.User) *cpb.BasicProfile {
	return &cpb.BasicProfile{
		UserId: src.ID.String,
		Name:   src.GetName(),
		Avatar: src.Avatar.String,
		Group:  cpb.UserGroup(cpb.UserGroup_value[src.Group.String]),
	}
}
