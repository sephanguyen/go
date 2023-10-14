package queries

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	infra_class "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentSubscriptionQueryHandler struct {
	WrapperConnection                 *support.WrapperDBConnection
	StudentSubscriptionRepo           infrastructure.StudentSubscriptionRepo
	StudentSubscriptionAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo
	ClassMemberRepo                   infra_class.ClassMemberRepo
	ClassRepo                         infra_class.ClassRepo
	Env                               string
	UnleashClientIns                  unleashclient.ClientInstance
}

type RetrieveStudentSubscriptionQueryResponse struct {
	Subs                       []*domain.StudentSubscription
	Total                      uint32
	PrePageID                  string
	SubsLocations              map[string][]string
	Classes                    []*class_domain.ClassWithCourseStudent
	IsEmptyStudentSubscription bool
	Err                        error
}

func (s *StudentSubscriptionQueryHandler) RetrieveStudentSubscription(ctx context.Context, args payloads.ListStudentSubScriptionsArgs) *RetrieveStudentSubscriptionQueryResponse {
	dbConn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &RetrieveStudentSubscriptionQueryResponse{
			Err: err,
		}
	}
	if len(args.ClassIDs) > 0 {
		studentIDWithCourseIDs, err := s.ClassMemberRepo.FindStudentIDWithCourseIDsByClassIDs(ctx, dbConn, args.ClassIDs)
		if err != nil {
			return &RetrieveStudentSubscriptionQueryResponse{
				Err: status.Error(codes.Internal, fmt.Errorf("cannot get ClassMemberRepo.FindStudentIDWithCourseIDsByClassIDs: %s", err).Error()),
			}
		}

		if len(studentIDWithCourseIDs) == 0 {
			return &RetrieveStudentSubscriptionQueryResponse{
				IsEmptyStudentSubscription: true,
			}
		}
		args.StudentIDWithCourseIDs = studentIDWithCourseIDs
	}
	if len(args.LocationIDs) > 0 {
		studentSubscriptionIDs, err := s.StudentSubscriptionAccessPathRepo.FindStudentSubscriptionIDsByLocationIDs(ctx, dbConn, args.LocationIDs)
		if err != nil {
			return &RetrieveStudentSubscriptionQueryResponse{
				Err: status.Error(codes.Internal, fmt.Errorf("cannot get StudentSubscriptionAccessPathRepo.FindStudentSubscriptionIDsByLocationIDs: %s", err).Error()),
			}
		}
		if len(studentSubscriptionIDs) == 0 {
			return &RetrieveStudentSubscriptionQueryResponse{
				IsEmptyStudentSubscription: true,
			}
		}
		args.StudentSubscriptionIDs = studentSubscriptionIDs
	}

	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return &RetrieveStudentSubscriptionQueryResponse{
			Err: status.Error(codes.InvalidArgument, "resource path is invalid"),
		}
	}

	args.SchoolID = fmt.Sprint(resourcePath)

	subs, total, prePageID, preTotal, err := s.StudentSubscriptionRepo.RetrieveStudentSubscription(ctx, dbConn, &args)
	if err != nil {
		return &RetrieveStudentSubscriptionQueryResponse{Err: status.Error(codes.Internal, err.Error())}
	}

	if preTotal <= args.Limit {
		prePageID = ""
	}
	studentSubscriptionIDs, studentCourses := getListStudentSubIDs(subs)
	var subsLocations map[string][]string
	var classes []*class_domain.ClassWithCourseStudent

	if len(studentSubscriptionIDs) > 0 {
		subsLocations, err = s.StudentSubscriptionAccessPathRepo.FindLocationsByStudentSubscriptionIDs(ctx, dbConn, studentSubscriptionIDs)
		if err != nil {
			return &RetrieveStudentSubscriptionQueryResponse{
				Err: status.Error(codes.Internal, fmt.Errorf("StudentSubscriptionAccessPathRepo.FindLocationsByStudentSubscriptionIDs: %w", err).Error()),
			}
		}

		classes, err = s.ClassRepo.FindByCourseIDsAndStudentIDs(ctx, dbConn, studentCourses)
		if err != nil {
			return &RetrieveStudentSubscriptionQueryResponse{
				Err: status.Error(codes.Internal, fmt.Errorf("ClassRepo.FindByCourseIDsAndStudentIDs: %w", err).Error()),
			}
		}
	}
	return &RetrieveStudentSubscriptionQueryResponse{
		Subs:          subs,
		Total:         total,
		PrePageID:     prePageID,
		SubsLocations: subsLocations,
		Classes:       classes,
		Err:           nil,
	}
}

func (s *StudentSubscriptionQueryHandler) GetStudentCourseSubscriptions(ctx context.Context, payload payloads.GetStudentCourseSubscriptions) (domain.StudentSubscriptions, []*class_domain.ClassWithCourseStudent, error) {
	var locationID []string
	dbConn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, nil, err
	}
	if len(payload.LocationID) > 0 {
		locationID = append(locationID, payload.LocationID)
	}
	studentSubscriptions, err := s.StudentSubscriptionRepo.GetStudentCourseSubscriptions(ctx, dbConn, locationID, payload.StudentIDWithCourseID...)
	classes := make([]*class_domain.ClassWithCourseStudent, 0)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf(`studentSubscriptionRepo.GetStudentCourseSubscriptions: %v`, err))
	}
	if err = studentSubscriptions.IsValid(); err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf(`invalid student subscription: %v`, err))
	}
	studentSubscriptionIDs, studentCourses := getListStudentSubIDs(studentSubscriptions)

	if len(studentSubscriptionIDs) > 0 {
		classes, err = s.ClassRepo.FindByCourseIDsAndStudentIDs(ctx, dbConn, studentCourses)
		if err != nil {
			return nil, nil, status.Error(codes.Internal, fmt.Sprintf(`studentSubscriptionRepo.GetStudentCourseSubscriptions FindByCourseIDsAndStudentIDs: %v`, err))
		}
	}
	return studentSubscriptions, classes, nil
}

func getListStudentSubIDs(subs []*domain.StudentSubscription) ([]string, []*class_domain.ClassWithCourseStudent) {
	studentSubscriptionIDs := make([]string, 0, len(subs))
	studentCourses := make([]*class_domain.ClassWithCourseStudent, 0, len(subs))
	for _, sub := range subs {
		studentSubscriptionIDs = append(studentSubscriptionIDs, sub.SubscriptionID)
		sc := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
		studentCourses = append(studentCourses, sc)
	}
	return studentSubscriptionIDs, studentCourses
}
