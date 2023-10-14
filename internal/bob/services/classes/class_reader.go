package classes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	opb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ClassReaderService struct {
	DB            database.Ext
	EurekaDBTrace database.Ext
	Env           string

	ClassRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) ([]*entities.Class, error)
	}
	LessonMemberRepo interface {
		ListStudentsByLessonID(ctx context.Context, db database.QueryExecer, args *repositories.ListStudentsByLessonArgs) ([]*entities.User, error)
	}
	ClassMemberRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.FindClassMemberFilter) ([]*entities.ClassMember, error)
		FindByClassIDsAndUserIDs(ctx context.Context, db database.QueryExecer, userIDs, classIDs pgtype.TextArray) ([]*entities.ClassMemberV2, error)
		FindByUserIDsAndCourseIDs(ctx context.Context, db database.QueryExecer, userIDs, courseIDs pgtype.TextArray) ([]*entities.ClassMemberV2, error)
	}
	UserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
	}
	StudentEventLogRepo interface {
		RetrieveStudentEventLogsByStudyPlanItemIDs(
			ctx context.Context, db database.QueryExecer, studyPlainItemIDs pgtype.TextArray,
		) ([]*entities.StudentEventLog, error)
	}

	UnleashClientIns unleashclient.ClientInstance
	MasterClassRepo  interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, classIDs []string) ([]*domain.Class, error)
	}
	MasterClassMemberRepo interface {
		RetrieveByClassIDs(ctx context.Context, db database.QueryExecer, filter *queries.FindClassMemberFilter) ([]*domain.ClassMember, error)
		RetrieveByClassMembers(ctx context.Context, db database.QueryExecer, filter *queries.RetrieveByClassMembersFilter) ([]string, error)
	}
	StudentEnrollmentHistory interface {
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, locationID pgtype.Text) ([]*entities.StudentEnrollmentStatusHistory, error)
	}
	LessonRepo interface {
		GetLessonByID(ctx context.Context, db database.QueryExecer, id string) (*lesson_domain.Lesson, error)
	}
	CourseReaderSvc interface {
		ListStudentIDsByCourse(context.Context, *epb.ListStudentIDsByCourseRequest, ...grpc.CallOption) (*epb.ListStudentIDsByCourseResponse, error)
		GetStudentsAccessPath(context.Context, *epb.GetStudentsAccessPathRequest, ...grpc.CallOption) (*epb.GetStudentsAccessPathResponse, error)
	}
	SchoolHistoryRepo interface {
		FindBySchoolAndStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs, schoolIDs pgtype.TextArray) ([]*entities.SchoolHistory, error)
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.SchoolHistory, error)
	}
	TaggedUserRepo interface {
		FindByTagIDsAndUserIDs(ctx context.Context, db database.QueryExecer, tagIDs, userIDs pgtype.TextArray) ([]*entities.TaggedUser, error)
	}
	StudyPlanReaderService interface {
		GetStudentStudyPlan(context.Context, *epb.GetStudentStudyPlanRequest, ...grpc.CallOption) (*epb.GetStudentStudyPlanResponse, error)
	}
}

func toClassPb(src *entities.Class) (*pb.Class, error) {
	id := strconv.Itoa(int(src.ID.Int))
	dst := &pb.Class{
		Id:           id,
		Name:         src.Name.String,
		Avatar:       src.Avatar.String,
		SchoolId:     src.SchoolID.Int,
		ClassCode:    src.Code.String,
		Subject:      nil,
		Grades:       nil,
		OwnerIds:     nil,
		TotalStudent: 0,
	}
	for _, s := range src.Subjects.Elements {
		dst.Subject = append(dst.Subject, cpb.Subject(cpb.Subject_value[s.String]))
	}

	for _, g := range src.Grades.Elements {
		grade, err := i18n.ConvertIntGradeToString(opb.Country(opb.Country_value[src.Country.String]), int(g.Int))
		if err != nil {
			return nil, errors.Wrap(err, "invalid class")
		}

		dst.Grades = append(dst.Grades, grade)
	}
	return dst, nil
}

func masterClassToClassPb(src *domain.Class) (*pb.Class, error) {
	dst := &pb.Class{
		Id:           src.ClassID,
		Name:         src.Name,
		SchoolId:     0,
		ClassCode:    "",
		Subject:      nil,
		Grades:       nil,
		OwnerIds:     nil,
		TotalStudent: 0,
	}

	return dst, nil
}

func (s *ClassReaderService) RetrieveClassByIDs(ctx context.Context, req *pb.RetrieveClassByIDsRequest) (*pb.RetrieveClassByIDsResponse, error) {
	if len(req.ClassIds) == 0 {
		return &pb.RetrieveClassByIDsResponse{}, nil
	}
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled("Architecture_BACKEND_RetrieveClassByIDs_Use_Mastermgmt_Repo", s.Env)
	if err != nil {
		isUnleashToggled = false
	}
	if isUnleashToggled {
		classes, err := s.MasterClassRepo.RetrieveByIDs(ctx, s.DB, req.ClassIds)
		if err != nil {
			return nil, fmt.Errorf("s.MasterClassRepo.FindByIDs: %w", err)
		}
		result := make([]*pb.Class, 0, len(classes))
		for _, class := range classes {
			classPb, err := masterClassToClassPb(class)
			if err != nil {
				return nil, fmt.Errorf("error convert masterClass toclassPb: %w", err)
			}
			result = append(result, classPb)
		}
		return &pb.RetrieveClassByIDsResponse{
			Classes: result,
		}, nil
	} else {
		ids := make([]int32, 0, len(req.ClassIds))
		for _, sID := range req.ClassIds {
			id, err := strconv.ParseInt(sID, 10, 32)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "")
			}
			ids = append(ids, int32(id))
		}
		classes, err := s.ClassRepo.FindByIDs(ctx, s.DB, database.Int4Array(ids))
		if err != nil {
			return nil, fmt.Errorf("s.ClassRepo.FindByIDs: %w", err)
		}
		result := make([]*pb.Class, 0, len(classes))
		for _, class := range classes {
			// doesn't return owners ids and total student as client don't need
			classPb, err := toClassPb(class)
			if err != nil {
				return nil, fmt.Errorf("error convert toclassPb: %w", err)
			}
			result = append(result, classPb)
		}

		return &pb.RetrieveClassByIDsResponse{
			Classes: result,
		}, nil
	}
}

func (s *ClassReaderService) ListClass(ctx context.Context, req *pb.ListClassRequest) (*pb.ListClassResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *ClassReaderService) RetrieveClassLearningStatistics(ctx context.Context, req *pb.RetrieveClassLearningStatisticsRequest) (*pb.RetrieveClassLearningStatisticsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *ClassReaderService) RetrieveClassMembers(ctx context.Context, req *pb.RetrieveClassMembersRequest) (*pb.RetrieveClassMembersResponse, error) {
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled("Architecture_BACKEND_RetrieveClassMembers_Use_Mastermgmt_Repo", s.Env)
	if err != nil {
		isUnleashToggled = false
	}
	var (
		data   []*pb.RetrieveClassMembersResponse_Member
		paging *cpb.Paging
	)
	if isUnleashToggled {
		// New table class_member (class_id text)
		data, paging, err = s.getClassMemberV2(ctx, req)
	} else {
		// Old table class_members (class_id int)
		data, paging, err = s.getClassMemberV1(ctx, req)
	}

	return &pb.RetrieveClassMembersResponse{
		Paging:  paging,
		Members: data,
	}, err
}

func (s *ClassReaderService) getClassMemberV2(ctx context.Context, req *pb.RetrieveClassMembersRequest) ([]*pb.RetrieveClassMembersResponse_Member, *cpb.Paging, error) {
	if len(req.ClassIds) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "invalid class ids")
	}
	filter := &queries.FindClassMemberFilter{
		ClassIDs: req.ClassIds,
		OffsetID: "",
		Limit:    0,
		UserName: "",
	}
	if req.Paging != nil {
		if req.Paging.GetOffsetString() != "" {
			filter.OffsetID = req.Paging.GetOffsetString()
		}
		if req.Paging.GetLimit() > 1 {
			filter.Limit = req.Paging.GetLimit()
		}
		if c := req.Paging.GetOffsetMultipleCombined(); c != nil {
			filter.UserName = c.GetCombined()[0].GetOffsetString()
			filter.OffsetID = c.GetCombined()[1].GetOffsetString()
		}
	}
	classMembers, err := s.MasterClassMemberRepo.RetrieveByClassIDs(ctx, s.DB, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("LessonMemberRepo.RetrieveByClassIDs: %w", err)
	}
	data := make([]*pb.RetrieveClassMembersResponse_Member, 0, len(classMembers))
	for _, m := range classMembers {
		joinedAt := timestamppb.New(m.CreatedAt)
		data = append(data, &pb.RetrieveClassMembersResponse_Member{
			UserId: m.UserID,
			JoinAt: joinedAt,
		})
	}
	if len(classMembers) == 0 {
		return data, nil, nil
	}
	usr, err := s.UserRepo.Get(ctx, s.DB, database.Text(classMembers[len(classMembers)-1].UserID))
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}
	paging := &cpb.Paging{
		Limit: req.Paging.GetLimit(),
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: usr.GetName(),
					},
					{
						OffsetString: classMembers[len(classMembers)-1].UserID,
					},
				},
			},
		},
	}
	return data, paging, nil
}

func (s *ClassReaderService) getClassMemberV1(ctx context.Context, req *pb.RetrieveClassMembersRequest) ([]*pb.RetrieveClassMembersResponse_Member, *cpb.Paging, error) {
	classIds, err := convertStringSliceToInt32Array(req.ClassIds)
	if err != nil || len(classIds) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "invalid class ids")
	}
	filter := &repositories.FindClassMemberFilter{
		ClassIDs: database.Int4Array(classIds),
		Status:   database.Text(entities.ClassMemberStatusActive),
	}
	if err := multierr.Combine(
		filter.Group.Set(nil),
		filter.OffsetID.Set(nil),
		filter.Limit.Set(nil),
		filter.UserName.Set(nil),
	); err != nil {
		return nil, nil, fmt.Errorf("RetrieveClassMembers.SetFilter: %w", err)
	}

	if req.Paging != nil {
		if req.Paging.GetOffsetString() != "" {
			_ = filter.OffsetID.Set(req.Paging.GetOffsetString())
		}
		if req.Paging.GetLimit() > 1 {
			_ = filter.Limit.Set(req.Paging.GetLimit())
		}
		if c := req.Paging.GetOffsetMultipleCombined(); c != nil {
			filter.UserName = database.Text(c.GetCombined()[0].GetOffsetString())
			filter.OffsetID = database.Text(c.GetCombined()[1].GetOffsetString())
		}
	}

	if req.UserGroup != cpb.UserGroup_USER_GROUP_NONE {
		_ = filter.Group.Set(req.UserGroup.String())
	}

	classMembers, err := s.ClassMemberRepo.Find(ctx, s.DB, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("LessonMemberRepo.Find: %w", err)
	}
	data := make([]*pb.RetrieveClassMembersResponse_Member, 0, len(classMembers))
	for _, m := range classMembers {
		joinedAt := timestamppb.New(m.CreatedAt.Time)
		data = append(data, &pb.RetrieveClassMembersResponse_Member{
			UserId:    m.UserID.String,
			UserGroup: cpb.UserGroup(cpb.UserGroup_value[m.UserGroup.String]),
			JoinAt:    joinedAt,
		})
	}
	if len(classMembers) == 0 {
		return data, nil, nil
	}
	usr, err := s.UserRepo.Get(ctx, s.DB, database.Text(classMembers[len(classMembers)-1].UserID.String))
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	paging := &cpb.Paging{
		Limit: req.Paging.GetLimit(),
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: usr.GetName(),
					},
					{
						OffsetString: classMembers[len(classMembers)-1].UserID.String,
					},
				},
			},
		},
	}
	return data, paging, nil
}

func (s *ClassReaderService) RetrieveStudentLearningStatistics(ctx context.Context, req *pb.RetrieveStudentLearningStatisticsRequest) (*pb.RetrieveStudentLearningStatisticsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *ClassReaderService) ListStudentsByLesson(ctx context.Context, req *pb.ListStudentsByLessonRequest) (*pb.ListStudentsByLessonResponse, error) {
	args := &repositories.ListStudentsByLessonArgs{
		LessonID: database.Text(req.LessonId),
		Limit:    10,
		UserName: pgtype.Text{Status: pgtype.Null},
		UserID:   pgtype.Text{Status: pgtype.Null},
	}
	if paging := req.Paging; paging != nil {
		if limit := paging.Limit; 1 <= limit && limit <= 100 {
			args.Limit = limit
		}
		if c := paging.GetOffsetMultipleCombined(); c != nil {
			args.UserName = database.Text(c.Combined[0].OffsetString)
			args.UserID = database.Text(c.Combined[1].OffsetString)
		}
	}

	students, err := s.LessonMemberRepo.ListStudentsByLessonID(ctx, s.DB, args)
	if err != nil {
		return nil, err
	}
	if len(students) == 0 {
		return &pb.ListStudentsByLessonResponse{}, nil
	}
	lesson, err := s.LessonRepo.GetLessonByID(ctx, s.DB, req.LessonId)
	if err != nil {
		return nil, err
	}
	studentID := make([]string, 0, len(students))
	for _, s := range students {
		studentID = append(studentID, s.ID.String)
	}
	studentEnrollment, err := s.StudentEnrollmentHistory.Retrieve(ctx, s.DB, database.TextArray(studentID), database.Text(lesson.LocationID))
	if err != nil {
		return nil, err
	}
	pbStudents := make([]*cpb.BasicProfile, 0, len(students))
	for _, student := range students {
		pbStudents = append(pbStudents, toCommonBasicProfile(student))
	}
	studentEnrollmentMap := make(map[string][]*entities.StudentEnrollmentStatusHistory, 0)
	for _, se := range studentEnrollment {
		studentEnrollmentMap[se.StudentID.String] = append(studentEnrollmentMap[se.StudentID.String], se)
	}
	lastItem := students[len(students)-1]
	nextPage := &cpb.Paging{
		Limit: args.Limit,
		Offset: &cpb.Paging_OffsetMultipleCombined{
			OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
				Combined: []*cpb.Paging_Combined{
					{
						OffsetString: lastItem.GetName(),
					},
					{
						OffsetString: lastItem.ID.String,
					},
				},
			},
		},
	}

	return &pb.ListStudentsByLessonResponse{
		Students:         pbStudents,
		NextPage:         nextPage,
		EnrollmentStatus: toEnrollmentStatus(studentEnrollmentMap),
	}, nil
}

func (s *ClassReaderService) RetrieveClassMembersWithFilters(ctx context.Context, req *pb.RetrieveClassMembersWithFiltersRequest) (*pb.RetrieveClassMembersWithFiltersResponse, error) {
	if req.CourseId == "" {
		return nil, status.Error(codes.InvalidArgument, "course_id can't be empty")
	}

	var limit uint32 = 100
	var offSetString string
	if req.Paging.GetLimit() > 0 {
		limit = req.Paging.GetLimit()
	}

	if req.Paging != nil {
		if req.Paging.GetOffsetString() != "" {
			offSetString = req.Paging.GetOffsetString()
		}
	}

	studentIDs, err := s.retrieveAllStudentIDsByCourseID(ctx, req.CourseId)
	if err != nil {

		return nil, status.Error(codes.Internal, fmt.Errorf("s.retrieveAllStudentIDsByCourseID: %w", err).Error())
	}

	listLocationIDsFilter := req.GetLocationIds()
	if len(listLocationIDsFilter) > 0 {
		studentIDs, err = s.retrieveStudentsAccessPathIDs(ctx, listLocationIDsFilter, studentIDs, []string{req.CourseId})

		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.retrieveStudentsAccessPathIDs: %w", err).Error())
		}
	}
	// filter by school
	allStudentIDsInSchool := []string{}
	allStudentIDsUnassignInSchool := []string{}

	if req.GetSchoolId() != "" || req.GetUnassigned() {
		if req.GetUnassigned() {
			allAssignStudentIDs := []string{}
			allAssignStudentsInCourse, err := s.SchoolHistoryRepo.FindByStudentIDs(ctx, s.DB, database.TextArray(studentIDs))
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("s.SchoolHistoryRepo.FindByStudentIDs: %w", err).Error())
			}
			for _, assignStudent := range allAssignStudentsInCourse {
				allAssignStudentIDs = append(allAssignStudentIDs, assignStudent.StudentID.String)
			}
			for _, studentID := range studentIDs {
				if !slices.Contains(allAssignStudentIDs, studentID) {
					allStudentIDsUnassignInSchool = append(allStudentIDsUnassignInSchool, studentID)
				}
			}

			studentIDs = allStudentIDsUnassignInSchool
		} else {
			schoolHistoryOfStudent, err := s.SchoolHistoryRepo.FindBySchoolAndStudentIDs(ctx, s.DB, database.TextArray([]string{req.GetSchoolId()}), database.TextArray(studentIDs))

			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Errorf("s.FindBySchoolAndStudentIDs: %w", err).Error())
			}
			for _, schoolHistory := range schoolHistoryOfStudent {
				allStudentIDsInSchool = append(allStudentIDsInSchool, schoolHistory.StudentID.String)
			}
			studentIDs = allStudentIDsInSchool
		}
	}

	// Get student_ids by class or unassign in this course
	listClassIDs := req.GetClassIds()
	if len(listClassIDs) > 0 {
		// get list students class of this course
		var classStudentIds []string
		var allUnAssignClassStudentIds []string
		classStudents, err := s.ClassMemberRepo.FindByClassIDsAndUserIDs(ctx, s.DB, database.TextArray(listClassIDs), database.TextArray(studentIDs))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassMemberRepo.FindByClassIDsAndUserIDs %w", err).Error())
		}

		for _, classStudent := range classStudents {
			classStudentIds = append(classStudentIds, classStudent.UserID.String)
		}

		classStudentIds = sliceutils.RemoveDuplicates(classStudentIds)

		indexOfUnAssignID := slices.Index(listClassIDs, constants.UnAssignClassID)
		if indexOfUnAssignID != -1 {
			assignStudentIDs := []string{}
			allAssignClassMembers, err := s.ClassMemberRepo.FindByUserIDsAndCourseIDs(ctx, s.DB, database.TextArray(studentIDs), database.TextArray([]string{req.CourseId}))
			if err != nil {
				return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassMemberRepo.FindByUserIDsAndCourseIDs %w", err).Error())
			}

			for _, classMember := range allAssignClassMembers {
				assignStudentIDs = append(assignStudentIDs, classMember.UserID.String)
			}

			for _, studentID := range studentIDs {
				if !slices.Contains(assignStudentIDs, studentID) {
					allUnAssignClassStudentIds = append(allUnAssignClassStudentIds, studentID)
				}
			}
			classStudentIds = append(classStudentIds, allUnAssignClassStudentIds...)
		}
		studentIDs = sliceutils.Intersect(studentIDs, classStudentIds)
	}
	// filter by tags
	tagIDs := req.GetStudentTagIds()
	if len(tagIDs) > 0 {
		// get list students class of this course
		var tagStudentIds []string
		classStudents, err := s.TaggedUserRepo.FindByTagIDsAndUserIDs(ctx, s.DB, database.TextArray(tagIDs), database.TextArray(studentIDs))
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ClassStudentRepo.FindByClassIDsAndUserIDs %w", err).Error())
		}

		for _, classStudent := range classStudents {
			tagStudentIds = append(tagStudentIds, classStudent.UserID.String)
		}
		studentIDs = sliceutils.RemoveDuplicates(tagStudentIds)
	}

	// filter by individual study plan and lo id
	if len(req.GetStudyPlanIds()) > 0 && len(req.GetLoIds()) > 0 {
		studentIDs, err = s.retrieveStudentIDsByStudyPlan(ctx, studentIDs, req.GetStudyPlanIds(), req.GetLoIds())
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.retrieveStudentIDsByStudyPlan: %w", err).Error())
		}
	}

	studentIDsPagination := []string{}

	if offSetString == "" {
		if len(studentIDs) >= int(limit) {
			studentIDsPagination = studentIDs[0:int(limit)]
		} else {
			studentIDsPagination = studentIDs
		}
	} else {
		offset := indexOf(offSetString, studentIDs)

		if offset != -1 {
			if len(studentIDs) >= offset+1+int(limit) {
				studentIDsPagination = studentIDs[offset+1 : offset+1+int(limit)]
			} else if len(studentIDs) >= offset+1 {
				studentIDsPagination = studentIDs[offset+1:]
			}
		}
	}

	resp := &pb.RetrieveClassMembersWithFiltersResponse{
		UserIds: studentIDsPagination,
		Paging: &cpb.Paging{
			Limit: limit,
		},
	}
	if len(studentIDsPagination) > 0 && len(studentIDsPagination) > int(limit) {
		resp.Paging.Offset = &cpb.Paging_OffsetString{
			OffsetString: studentIDsPagination[len(studentIDsPagination)-1],
		}
	}
	return resp, nil
}

func (s *ClassReaderService) retrieveAllStudentIDsByCourseID(ctx context.Context, courseID string) (studentIDs []string, _ error) {
	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("interceptors.GetOutgoingContext: %w", err).Error())
	}

	limit := 2000

	req := &epb.ListStudentIDsByCourseRequest{
		CourseIds: []string{courseID},
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
	}

	for {
		res, err := s.CourseReaderSvc.ListStudentIDsByCourse(cctx, req)
		if err != nil {
			return nil, err
		}

		for _, sc := range res.GetStudentCourses() {
			studentIDs = append(studentIDs, sc.GetStudentId())
		}

		// Break
		length := len(res.GetStudentCourses())
		if res.GetNextPage() == nil || res.GetNextPage().GetOffset() == nil || length == 0 || length < limit {
			break
		}

		req.Paging = res.NextPage
	}

	return studentIDs, nil
}

func toCommonBasicProfile(e *entities.User) *cpb.BasicProfile {
	basicProfile := &cpb.BasicProfile{
		UserId:      e.ID.String,
		Name:        e.GetName(),
		Avatar:      e.Avatar.String,
		Group:       cpb.UserGroup(cpb.UserGroup_value[e.Group.String]),
		FacebookId:  e.FacebookID.String,
		AppleUserId: e.AppleUser.ID.String,
	}
	if !e.LastLoginDate.Time.IsZero() {
		basicProfile.LastLoginDate = timestamppb.New(e.LastLoginDate.Time)
	}
	return basicProfile
}

func toEnrollmentStatus(studentEnrollmentMap map[string][]*entities.StudentEnrollmentStatusHistory) []*pb.EnrollmentStatus {
	enrollmentStatus := make([]*pb.EnrollmentStatus, 0)
	for studentID, sem := range studentEnrollmentMap {
		en := &pb.EnrollmentStatus{
			StudentId: studentID,
			Info: func() (res []*pb.EnrollmentStatus_EnrollmentStatusInfo) {
				for _, s := range sem {
					res = append(res, &pb.EnrollmentStatus_EnrollmentStatusInfo{
						LocationId: s.LocationID.String,
						StartDate:  timestamppb.New(s.StartDate.Time),
						EndDate:    timestamppb.New(s.EndDate.Time),
					})
				}
				return
			}(),
		}
		enrollmentStatus = append(enrollmentStatus, en)
	}

	return enrollmentStatus
}

func convertStringSliceToInt32Array(ss []string) ([]int32, error) {
	result := make([]int32, 0, len(ss))
	for _, element := range ss {
		val, err := strconv.ParseInt(element, 10, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, int32(val))
	}
	return result, nil
}

func (s *ClassReaderService) retrieveStudentsAccessPathIDs(ctx context.Context, locationIDs, reqStudentIDs, courseIDs []string) (studentIDs []string, _ error) {
	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("interceptors.GetOutgoingContext: %w", err).Error())
	}

	req := &epb.GetStudentsAccessPathRequest{
		LocationIds: locationIDs,
		StudentIds:  reqStudentIDs,
		CourseIds:   courseIDs,
	}

	res, err := s.CourseReaderSvc.GetStudentsAccessPath(cctx, req)
	if err != nil {
		return nil, err
	}
	resStudentIDs := []string{}

	for _, sc := range res.GetCourseStudentAccesssPaths() {
		if !slices.Contains(resStudentIDs, sc.GetStudentId()) {
			resStudentIDs = append(resStudentIDs, sc.GetStudentId())
		}
	}

	return resStudentIDs, nil
}

func (s *ClassReaderService) retrieveStudentIDsByStudyPlan(ctx context.Context, reqStudentIDs, studyPlanIDs, loIDs []string) ([]string, error) {
	cctx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("interceptors.GetOutgoingContext: %w", err).Error())
	}

	req := &epb.GetStudentStudyPlanRequest{
		StudyPlanIds:        studyPlanIDs,
		StudentIds:          reqStudentIDs,
		LearningMaterialIds: loIDs,
	}

	res, err := s.StudyPlanReaderService.GetStudentStudyPlan(cctx, req)
	if err != nil {
		return nil, err
	}
	resStudentIDs := []string{}
	for _, sc := range res.GetStudentStudyPlans() {
		if !slices.Contains(resStudentIDs, sc.GetStudentId()) {
			resStudentIDs = append(resStudentIDs, sc.GetStudentId())
		}
	}

	return resStudentIDs, nil

}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}
