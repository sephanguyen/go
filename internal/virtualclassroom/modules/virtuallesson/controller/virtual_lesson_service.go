package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type VirtualLessonReaderService struct {
	WrapperDBConnection *support.WrapperDBConnection
	JSM                 nats.JetStreamManagement
	Env                 string

	VirtualLessonQuery queries.VirtualLessonQuery

	VirtualLessonRepo infrastructure.VirtualLessonRepo
	LessonGroupRepo   infrastructure.LessonGroupRepo

	UnleashClient unleashclient.ClientInstance
}

func (v *VirtualLessonReaderService) GetVirtualLessonByID(ctx context.Context, lessonID string, opts ...GetVirtualLessonOption) (*domain.VirtualLesson, error) {
	// lesson and list its teachers, learners
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	lesson, err := v.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.FindByID: %s", err)
	}

	for _, opt := range opts {
		if err = opt(ctx, v, lesson); err != nil {
			return nil, err
		}
	}

	return lesson, nil
}

type GetVirtualLessonOption func(context.Context, *VirtualLessonReaderService, *domain.VirtualLesson) error

func IncludeLearnerIDs() GetVirtualLessonOption {
	return func(ctx context.Context, s *VirtualLessonReaderService, lesson *domain.VirtualLesson) error {
		conn, err := s.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return err
		}
		learnerIDs, err := s.VirtualLessonRepo.GetLearnerIDsOfLesson(ctx, conn, lesson.LessonID)
		if err != nil {
			return fmt.Errorf("LessonRepo.GetLearnerIDsOfLesson: %s", err)
		}
		lesson.LearnerIDs.LearnerIDs = learnerIDs

		return nil
	}
}

func IncludeTeacherIDs() GetVirtualLessonOption {
	return func(ctx context.Context, s *VirtualLessonReaderService, lesson *domain.VirtualLesson) error {
		conn, err := s.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return err
		}
		teacherIDs, err := s.VirtualLessonRepo.GetTeacherIDsOfLesson(ctx, conn, lesson.LessonID)
		if err != nil {
			return fmt.Errorf("LessonRepo.GetTeacherIDsOfLesson: %s", err)
		}
		lesson.TeacherIDs.TeacherIDs = teacherIDs

		return nil
	}
}

func (v *VirtualLessonReaderService) GetLiveLessonsByLocations(ctx context.Context, req *vpb.GetLiveLessonsByLocationsRequest) (*vpb.GetLiveLessonsByLocationsResponse, error) {
	payload := convertLiveLessonRequestToPayload(req)

	isUnleashToggledWhitelist, err := v.UnleashClient.IsFeatureEnabled("VirtualClassroom_Whitelist_CourseIDs_Get_Live_Lesson", v.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash, whitelist get live lessons: %w", err)
	}
	payload.GetWhitelistCourseIDs = isUnleashToggledWhitelist

	response, err := v.VirtualLessonQuery.GetLiveLessonsByLocations(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	lessonsPb := toLessonsPb(response)

	return &vpb.GetLiveLessonsByLocationsResponse{
		Lessons: lessonsPb,
		Total:   response.Total,
	}, nil
}

func (v *VirtualLessonReaderService) GetLearnersByLessonID(ctx context.Context, req *vpb.GetLearnersByLessonIDRequest) (*vpb.GetLearnersByLessonIDResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}

	payload := convertLearnersByLessonIDRequestToPayload(req)
	response, err := v.VirtualLessonQuery.GetLearnersByLessonID(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(response.StudentIDs) == 0 {
		return &vpb.GetLearnersByLessonIDResponse{}, nil
	}

	nextPage := &cpb.Paging{
		Limit: uint32(response.Limit),
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: response.LastLessonCourseID,
					},
					{
						OffsetString: response.LastUserID,
					},
				},
			},
		},
	}

	return &vpb.GetLearnersByLessonIDResponse{
		Learners: toLearnerInfoPb(response.StudentIDs, response.StudentInfo),
		NextPage: nextPage,
	}, nil
}

func (v *VirtualLessonReaderService) GetLearnersByLessonIDs(ctx context.Context, req *vpb.GetLearnersByLessonIDsRequest) (*vpb.GetLearnersByLessonIDsResponse, error) {
	if len(req.GetLessonId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "lesson IDs cannot be empty, at least 1 lesson ID is required")
	}

	lessonLearnersMap, err := v.VirtualLessonQuery.GetLearnersByLessonIDs(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(lessonLearnersMap) == 0 {
		return &vpb.GetLearnersByLessonIDsResponse{}, nil
	}

	return &vpb.GetLearnersByLessonIDsResponse{
		LessonLearners: toLessonLearnersPb(lessonLearnersMap),
	}, nil
}

func (v *VirtualLessonReaderService) GetClassDoURL(ctx context.Context, req *vpb.GetClassDoURLRequest) (*vpb.GetClassDoURLResponse, error) {
	if len(req.GetLessonId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "lesson ID cannot be empty")
	}

	classDoLink, err := v.VirtualLessonQuery.GetClassDoInfoByLessonID(ctx, req.GetLessonId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.GetClassDoURLResponse{
		ClassdoLink: classDoLink,
	}, nil
}

func convertLiveLessonRequestToPayload(req *vpb.GetLiveLessonsByLocationsRequest) *payloads.GetLiveLessonsByLocationsRequest {
	payload := &payloads.GetLiveLessonsByLocationsRequest{
		LocationIDs: req.GetLocationIds(),
		CourseIDs:   req.GetCourseIds(),
		EndDate:     req.GetTo().AsTime(),
	}

	if startDate := req.GetFrom(); startDate != nil {
		payload.StartDate = startDate.AsTime()
	}

	if endDate := req.GetTo(); endDate != nil {
		payload.EndDate = endDate.AsTime()
	}

	if statuses := req.GetSchedulingStatus(); len(statuses) > 0 {
		lessonStatuses := make([]domain.LessonSchedulingStatus, 0, len(statuses))
		for _, status := range statuses {
			lessonStatuses = append(lessonStatuses, domain.LessonSchedulingStatus(status.String()))
		}
		payload.LessonSchedulingStatuses = lessonStatuses
	}

	if pagination := req.GetPagination(); pagination != nil {
		payload.Limit = pagination.GetLimit()
		payload.Page = pagination.GetPage()
	}

	return payload
}

func convertLearnersByLessonIDRequestToPayload(req *vpb.GetLearnersByLessonIDRequest) *payloads.GetLearnersByLessonIDArgs {
	payload := &payloads.GetLearnersByLessonIDArgs{
		LessonID: req.GetLessonId(),
		Limit:    10,
	}

	if paging := req.GetPaging(); paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			payload.Limit = int32(limit)
		}

		if c := paging.GetOffsetMultipleCombined(); c != nil {
			payload.LessonCourseID = c.Combined[0].OffsetString
			payload.UserID = c.Combined[1].OffsetString
		}
	}

	return payload
}

func toLessonsPb(res *payloads.GetLiveLessonsByLocationsResponse) []*vpb.Lesson {
	lessonsPb := make([]*vpb.Lesson, 0, len(res.Lessons))

	for _, lesson := range res.Lessons {
		status := cpb.LessonStatus_LESSON_STATUS_NOT_STARTED
		now := timeutil.Now().Unix()

		if lesson.StartTime.Unix() >= now {
			status = cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS
		}
		if lesson.EndTime.Unix() < now || lesson.EndAt != nil {
			status = cpb.LessonStatus_LESSON_STATUS_COMPLETED
		}

		lessonPb := &vpb.Lesson{
			LessonId:                 lesson.LessonID,
			CourseId:                 lesson.CourseID,
			PresetStudyPlanWeeklyIds: "",
			Topic: &cpb.Topic{
				Attachments: []*cpb.Attachment{},
			},
			StartTime:      timestamppb.New(lesson.StartTime),
			EndTime:        timestamppb.New(lesson.EndTime),
			Status:         status,
			ZoomLink:       lesson.ZoomLink,
			TeachingMedium: cpb.LessonTeachingMedium(cpb.LessonTeachingMedium_value[string(lesson.TeachingMedium)]),
		}

		if len(lesson.TeacherID) > 0 {
			lessonPb.Teacher = []*cpb.BasicProfile{
				{
					UserId: lesson.TeacherID,
				},
			}
		}

		lessonsPb = append(lessonsPb, lessonPb)
	}

	return lessonsPb
}

func toLearnerInfoPb(learnerIDs []string, learnerInfoMap map[string][]*domain.StudentEnrollmentStatusHistory) []*vpb.LearnerInfo {
	learnerInfosPb := make([]*vpb.LearnerInfo, 0, len(learnerIDs))

	for _, learnerID := range learnerIDs {
		enrollmentInfos := learnerInfoMap[learnerID]

		learnerInfoPb := &vpb.LearnerInfo{
			LearnerId: learnerID,
			EnrollmentStatusInfo: func() (res []*vpb.LearnerInfo_EnrollmentStatusInfo) {
				for _, enrollmentInfo := range enrollmentInfos {
					res = append(res, &vpb.LearnerInfo_EnrollmentStatusInfo{
						LocationId: enrollmentInfo.LocationID,
						StartDate:  timestamppb.New(enrollmentInfo.StartDate),
						EndDate:    timestamppb.New(enrollmentInfo.EndDate),
					})
				}
				return
			}(),
		}
		learnerInfosPb = append(learnerInfosPb, learnerInfoPb)
	}

	return learnerInfosPb
}

func toLessonLearnersPb(ll map[string]domain.LessonLearners) []*vpb.GetLearnersByLessonIDsResponse_LessonLearners {
	lessonLearnersPb := make([]*vpb.GetLearnersByLessonIDsResponse_LessonLearners, 0, len(ll))

	for lessonID, learners := range ll {
		lessonLearner := &vpb.GetLearnersByLessonIDsResponse_LessonLearners{
			LessonId: lessonID,
			Learners: func() (res []*vpb.GetLearnersByLessonIDsResponse_LessonLearners_Learner) {
				for _, learner := range learners {
					res = append(res, &vpb.GetLearnersByLessonIDsResponse_LessonLearners_Learner{
						LearnerId: learner.LearnerID,
					})
				}
				return
			}(),
		}
		lessonLearnersPb = append(lessonLearnersPb, lessonLearner)
	}

	return lessonLearnersPb
}

func (v *VirtualLessonReaderService) GetLessons(ctx context.Context, req *vpb.GetLessonsRequest) (*vpb.GetLessonsResponse, error) {
	var lastItem string

	if req.GetPaging() == nil {
		return nil, status.Error(codes.InvalidArgument, "missing paging info")
	}
	if req.GetCurrentTime() == nil {
		return nil, status.Error(codes.InvalidArgument, "missing current time")
	}

	payload := getPayloadFromRequest(ctx, req)
	lessons, total, offsetID, err := v.VirtualLessonQuery.GetLessons(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	lessonLength := len(lessons)
	items := make([]*vpb.GetLessonsResponse_Lesson, 0, lessonLength)
	for _, lesson := range lessons {
		items = append(items, toGetLessonsResponseLessonPb(lesson))
	}
	if lessonLength > 0 {
		lastItem = lessons[lessonLength-1].LessonID
	}

	return &vpb.GetLessonsResponse{
		Items:       items,
		TotalItems:  total,
		TotalLesson: total,
		NextPage: &cpb.Paging{
			Limit: payload.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: payload.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: offsetID,
			},
		},
	}, nil
}

func getPayloadFromRequest(ctx context.Context, req *vpb.GetLessonsRequest) payloads.GetLessonsArgs {
	paging := req.GetPaging()
	filter := req.GetFilter()

	args := payloads.GetLessonsArgs{
		CurrentTime:      req.CurrentTime.AsTime(),
		TimeLookup:       payloads.TimeLookup(req.GetTimeLookup().String()),
		SortAscending:    req.GetSortAsc(),
		SchoolID:         golibs.ResourcePathFromCtx(ctx),
		Limit:            paging.GetLimit(),
		OffsetLessonID:   paging.GetOffsetString(),
		LocationIDs:      req.GetLocationIds(),
		TeacherIDs:       filter.GetTeacherIds(),
		StudentIDs:       filter.GetStudentIds(),
		CourseIDs:        filter.GetCourseIds(),
		LiveLessonStatus: payloads.LiveLessonStatus(filter.GetLiveLessonStatus().String()),
	}

	switch req.GetLessonTimeCompare() {
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE:
		args.LessonTimeCompare = payloads.LessonTimeCompareFuture
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE_AND_EQUAL:
		args.LessonTimeCompare = payloads.LessonTimeCompareFutureAndEqual
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_PAST:
		args.LessonTimeCompare = payloads.LessonTimeComparePast
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_PAST_AND_EQUAL:
		args.LessonTimeCompare = payloads.LessonTimeComparePastAndEqual
	}

	if statuses := filter.GetSchedulingStatus(); len(statuses) > 0 {
		lessonStatuses := make([]domain.LessonSchedulingStatus, 0, len(statuses))
		for _, status := range statuses {
			lessonStatuses = append(lessonStatuses, domain.LessonSchedulingStatus(status.String()))
		}
		args.LessonSchedulingStatuses = lessonStatuses
	}

	if fromDate := filter.GetFromDate(); fromDate != nil {
		args.FromDate = fromDate.AsTime()
	}

	if toDate := filter.GetToDate(); toDate != nil {
		args.ToDate = toDate.AsTime()
	}

	return args
}

func toGetLessonsResponseLessonPb(l domain.VirtualLesson) *vpb.GetLessonsResponse_Lesson {
	return &vpb.GetLessonsResponse_Lesson{
		Id:               l.LessonID,
		Name:             l.Name,
		CenterId:         l.CenterID,
		StartTime:        timestamppb.New(l.StartTime),
		EndTime:          timestamppb.New(l.EndTime),
		TeacherIds:       l.TeacherIDs.TeacherIDs,
		TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[string(l.TeachingMethod)]),
		TeachingMedium:   cpb.LessonTeachingMedium(cpb.LessonTeachingMedium_value[string(l.TeachingMedium)]),
		CourseId:         l.CourseID,
		ClassId:          l.ClassID,
		SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(l.SchedulingStatus)]),
		LessonCapacity:   uint32(l.LessonCapacity),
		EndAt: func() *timestamppb.Timestamp {
			if l.EndAt == nil {
				return nil
			}
			return timestamppb.New(*l.EndAt)
		}(),
		ZoomLink: l.ZoomLink,
	}
}
