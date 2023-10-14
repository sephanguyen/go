package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	infra_class "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StudentSubscriptionGRPCService struct {
	StudentSubScription     queries.StudentSubscriptionQueryHandler
	StudentReallocate       queries.StudentReallocateQueryHandler
	studentSubscriptionRepo infrastructure.StudentSubscriptionRepo
	wrapperConnection       *support.WrapperDBConnection
	UnleashClient           unleashclient.ClientInstance
	Env                     string
}

func NewStudentSubscriptionGRPCService(wrapperConnection *support.WrapperDBConnection,
	studentSubscriptionRepo infrastructure.StudentSubscriptionRepo,
	studentSubscriptionAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo,
	classMemberRepo infra_class.ClassMemberRepo,
	classRepo infra_class.ClassRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,

) *StudentSubscriptionGRPCService {
	studentSub := queries.StudentSubscriptionQueryHandler{
		WrapperConnection:                 wrapperConnection,
		StudentSubscriptionRepo:           studentSubscriptionRepo,
		StudentSubscriptionAccessPathRepo: studentSubscriptionAccessPathRepo,
		ClassMemberRepo:                   classMemberRepo,
		ClassRepo:                         classRepo,
		UnleashClientIns:                  unleashClientIns,
		Env:                               env,
	}
	studentReallocate := queries.StudentReallocateQueryHandler{
		WrapperConnection:       wrapperConnection,
		StudentSubscriptionRepo: studentSubscriptionRepo,
		UnleashClientIns:        unleashClientIns,
		Env:                     env,
	}
	return &StudentSubscriptionGRPCService{
		StudentSubScription:     studentSub,
		StudentReallocate:       studentReallocate,
		studentSubscriptionRepo: studentSubscriptionRepo,
		wrapperConnection:       wrapperConnection,
		UnleashClient:           unleashClientIns,
		Env:                     env,
	}
}

func (s *StudentSubscriptionGRPCService) GetStudentCourseSubscriptions(ctx context.Context, req *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
	studentIDWithCourseID := make([]string, 0, len(req.Subscriptions)*2)
	for _, sub := range req.Subscriptions {
		studentIDWithCourseID = append(studentIDWithCourseID, sub.StudentId, sub.CourseId)
	}

	payload := payloads.GetStudentCourseSubscriptions{
		LocationID:            req.LocationId,
		StudentIDWithCourseID: studentIDWithCourseID,
	}
	studentSubscriptions, classes, err := s.StudentSubScription.GetStudentCourseSubscriptions(ctx, payload)
	if err != nil {
		return nil, err
	}

	res := &lpb.GetStudentCourseSubscriptionsResponse{}
	res.Items = make([]*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription, 0, len(studentSubscriptions))
	for _, sub := range studentSubscriptions {
		classID := ""
		for _, v := range classes {
			if v.CourseID == sub.CourseID && v.StudentID == sub.StudentID {
				classID = v.ClassID
				break
			}
		}
		res.Items = append(res.Items, &lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
			Id:          sub.SubscriptionID,
			StudentId:   sub.StudentID,
			CourseId:    sub.CourseID,
			ClassId:     classID,
			LocationIds: sub.LocationIDs,
			StartDate:   timestamppb.New(sub.StartAt),
			EndDate:     timestamppb.New(sub.EndAt),
			GradeV2:     sub.GradeV2,
		})
	}
	return res, nil
}

func (s *StudentSubscriptionGRPCService) RetrieveStudentSubscription(ctx context.Context, req *lpb.RetrieveStudentSubscriptionRequest) (*lpb.RetrieveStudentSubscriptionResponse, error) {
	if req.GetPaging() == nil {
		return nil, status.Error(codes.Internal, "missing paging info")
	}
	args := filterArgsSubFromRequest(req)
	if req.GetPaging().GetOffsetString() != "" {
		args.StudentSubscriptionID = req.Paging.GetOffsetString()
	}

	result := s.StudentSubScription.RetrieveStudentSubscription(ctx, args)
	if result.IsEmptyStudentSubscription {
		return &lpb.RetrieveStudentSubscriptionResponse{
			Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{},
			NextPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			TotalItems: 0,
		}, nil
	}

	if result.Err != nil {
		return nil, result.Err
	}

	items := []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{}
	for i, sub := range result.Subs {
		locationIDs := result.SubsLocations[sub.SubscriptionID]
		items = append(items, toStudentSubs(sub, locationIDs))

		for _, v := range result.Classes {
			if v.CourseID == sub.CourseID && v.StudentID == sub.StudentID {
				items[i].ClassId = v.ClassID
				break
			}
		}
	}
	lastItem := ""
	if len(result.Subs) > 0 {
		lastItem = result.Subs[len(result.Subs)-1].SubscriptionID
	}
	return &lpb.RetrieveStudentSubscriptionResponse{
		Items: items,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: result.PrePageID,
			},
		},
		TotalItems: result.Total,
	}, nil
}

func (s *StudentSubscriptionGRPCService) RetrieveStudentPendingReallocate(ctx context.Context, req *lpb.RetrieveStudentPendingReallocateRequest) (*lpb.RetrieveStudentPendingReallocateResponse, error) {
	limit := req.Paging.GetLimit()
	offset := req.Paging.GetOffsetInteger()
	timezone := req.GetTimezone()
	if len(timezone) == 0 {
		timezone = "Asia/Ho_Chi_Minh"
	}
	loc, _ := time.LoadLocation(timezone)
	parserResp, err := s.StudentReallocate.RetrieveStudentPendingReallocate(ctx, queries.StudentReallocateRequest{
		SearchKey:  req.Keyword,
		LessonDate: req.LessonDate.AsTime(),
		Filter: domain.Filter{
			CourseID:   req.Filter.GetCourseId(),
			LocationID: req.Filter.GetLocationId(),
			GradeID:    req.Filter.GetGradeId(),
			ClassId:    req.Filter.GetClassId(),
			StartDate:  support.StartOfDate(req.Filter.GetStartDate().AsTime().In(loc)),
			EndDate:    support.EndOfDate(req.Filter.GetEndDate().AsTime().In(loc)),
		},
		Paging: support.Paging[int]{
			Limit:  int(limit),
			Offset: int(offset),
		},
		Timezone: timezone,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*lpb.RetrieveStudentPendingReallocateResponse_ReallocateStudent, 0, len(parserResp.ReallocateStudent))
	for _, rs := range parserResp.ReallocateStudent {
		items = append(items, &lpb.RetrieveStudentPendingReallocateResponse_ReallocateStudent{
			StudentId:        rs.StudentId,
			OriginalLessonId: rs.OriginalLessonID,
			GradeId:          rs.GradeID,
			CourseId:         rs.CourseID,
			LocationId:       rs.LocationID,
			ClassId:          rs.ClassID,
			StartDate:        timestamppb.New(rs.StartAt),
			EndDate:          timestamppb.New(rs.EndAt),
		})
	}
	preOffset := offset
	if offset > 0 {
		preOffset = offset - int64(limit)
	}
	apiResp := &lpb.RetrieveStudentPendingReallocateResponse{
		Items: items,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: offset + int64(limit),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: preOffset,
			},
		},
		TotalItems: parserResp.Total,
	}
	return apiResp, nil
}

func filterArgsSubFromRequest(req *lpb.RetrieveStudentSubscriptionRequest) payloads.ListStudentSubScriptionsArgs {
	args := payloads.ListStudentSubScriptionsArgs{
		Limit: req.Paging.Limit,
	}

	if courses := req.GetFilter().GetCourseId(); len(courses) > 0 {
		args.CourseIDs = courses
	}

	if keyWord := req.GetKeyword(); keyWord != "" {
		args.KeyWord = keyWord
	}

	if grades := req.GetFilter().GetGrade(); len(grades) > 0 {
		ints := []int32{}
		for _, grade := range grades {
			j, err := strconv.ParseInt(grade, 10, 32)
			if err != nil {
				panic(err)
			}
			ints = append(ints, int32(j))
		}
		args.Grades = ints
	}

	if gradesV2 := req.GetFilter().GetGradesV2(); len(gradesV2) > 0 {
		args.GradesV2 = gradesV2
	}

	if classes := req.GetFilter().GetClassId(); len(classes) > 0 {
		args.ClassIDs = classes
	}

	if locations := req.GetFilter().GetLocationId(); len(locations) > 0 {
		args.LocationIDs = locations
	}

	if lessonDate := req.GetLessonDate(); lessonDate != nil {
		args.LessonDate = lessonDate.AsTime()
	}

	return args
}

func toStudentSubs(s *domain.StudentSubscription, locationIDs []string) *lpb.RetrieveStudentSubscriptionResponse_StudentSubscription {
	return &lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
		Id:          s.SubscriptionID,
		StudentId:   s.StudentID,
		CourseId:    s.CourseID,
		Grade:       s.Grade,
		GradeV2:     s.GradeV2,
		LocationIds: locationIDs,
		StartDate:   timestamppb.New(s.StartAt),
		EndDate:     timestamppb.New(s.EndAt),
	}
}

func (s *StudentSubscriptionGRPCService) GetStudentCoursesAndClasses(ctx context.Context, req *lpb.GetStudentCoursesAndClassesRequest) (*lpb.GetStudentCoursesAndClassesResponse, error) {
	conn, err := s.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	data, err := s.studentSubscriptionRepo.GetStudentCoursesAndClasses(ctx, conn, req.GetStudentId())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("studentSubscriptionRepo.GetStudentCoursesAndClasses: %v", err))
	}
	return data.ToGetStudentCoursesAndClassesResponse(), nil
}
