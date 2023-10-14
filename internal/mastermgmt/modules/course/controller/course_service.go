package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	course_infras "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterDataCourseService struct {
	DB                                database.Ext
	LocationRepo                      infrastructure.LocationRepo
	CourseTypeRepo                    course_infras.CourseTypeRepo
	StudentSubscriptionCommandHandler queries.StudentSubscriptionQueryHandler
	CourseCommandHandler              commands.CourseCommandHandler
	CourseQueryHandler                queries.CourseQueryHandler
	UnleashClientIns                  unleashclient.ClientInstance
	Env                               string
}

func (m *MasterDataCourseService) ExportCourses(ctx context.Context, req *mpb.ExportCoursesRequest) (res *mpb.ExportCoursesResponse, err error) {
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	enableTeachingMethod, err := m.UnleashClientIns.IsFeatureEnabledOnOrganization("Architecture_BACKEND_MasterData_Course_TeachingMethod", m.Env, resourcePath)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can not connect unleash, %v", err.Error())
	}
	bytes, err := m.CourseQueryHandler.ExportCourses(ctx, enableTeachingMethod)
	if err != nil {
		return &mpb.ExportCoursesResponse{}, err
	}
	res = &mpb.ExportCoursesResponse{
		Data: bytes,
	}
	return res, nil
}

func (m *MasterDataCourseService) ImportCourses(ctx context.Context, req *mpb.ImportCoursesRequest) (res *mpb.ImportCoursesResponse, err error) {
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	enableTeachingMethod, err := m.UnleashClientIns.IsFeatureEnabledOnOrganization("Architecture_BACKEND_MasterData_Course_TeachingMethod", m.Env, resourcePath)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can not connect unleash, %v", err.Error())
	}

	config := validators.CSVImportConfig[domain.Course]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "course_id",
				Required: false,
			},
			{
				Column:   "course_name",
				Required: true,
			},
			{
				Column:   "course_type_id",
				Required: false,
			},
			{
				Column:   "course_partner_id",
				Required: false,
			},
			{
				Column:   "remarks",
				Required: false,
			},
		},
		Transform: transformCSVLineToCourse,
	}
	if enableTeachingMethod {
		config = validators.CSVImportConfig[domain.Course]{
			ColumnConfig: []validators.CSVColumn{
				{
					Column:   "course_id",
					Required: false,
				},
				{
					Column:   "course_name",
					Required: true,
				},
				{
					Column:   "course_type_id",
					Required: false,
				},
				{
					Column:   "course_partner_id",
					Required: false,
				},
				{
					Column:   "teaching_method",
					Required: false,
				},
				{
					Column:   "remarks",
					Required: false,
				},
			},
			Transform: transformCSVLineToCourseV2,
		}
	}
	csvCourses, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	courseTypeIDs := sliceutils.Map(csvCourses, mapToCourseTypeIDs)
	// check course type
	courseTypes, err := m.CourseTypeRepo.GetByIDs(ctx, m.DB, courseTypeIDs)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not validate course type id, %v", err.Error())
	}
	for k, v := range csvCourses {
		// skip if there are syntax errors or empty course type id
		if v.Value.CourseTypeID != "" && v.Error == nil {
			isCourseTypeExist := sliceutils.ContainFunc(courseTypes, &domain.CourseType{CourseTypeID: v.Value.CourseTypeID}, compareByCourseTypeIDs)
			if !isCourseTypeExist {
				v.Error = &dto.UpsertError{
					RowNumber: int32(k + 2),
					Error:     fmt.Sprintf("course type id %s is not exist", v.Value.CourseTypeID),
				}
			}
		}
	}

	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvCourses, validators.GetErrorFromCSVValue[domain.Course], validators.HasCSVErr[domain.Course])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	resourcePathInt, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	courses := sliceutils.Map(csvCourses, mapCourseCSVtoCourse(resourcePathInt))
	payload := commands.ImportCoursesPayload{
		Courses: courses,
	}
	err = m.CourseCommandHandler.ImportCourses(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportCoursesResponse{}, nil
}

func (m *MasterDataCourseService) UpsertCourses(ctx context.Context, req *mpb.UpsertCoursesRequest) (*mpb.UpsertCoursesResponse, error) {
	courseIDs := make([]string, 0, len(req.Courses))
	domainCourses := make([]*domain.Course, 0, len(req.Courses))
	courseAccessPaths := []*domain.CourseAccessPath{}
	for _, c := range req.Courses {
		resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
		}
		builder := domain.NewCourse().
			WithLocationRepo(m.LocationRepo).
			WithCourseTypeRepo(m.CourseTypeRepo).
			WithCourseID(c.Id).
			WithName(c.Name).
			WithIcon(c.Icon).
			WithDisplayOrder(int(c.DisplayOrder)).
			WithLocationIDs(c.LocationIds).
			WithSchoolID(int(resourcePath)).
			WithCourseType(c.CourseType).
			WithTeachingMethod(domain.CourseTeachingMethod(c.TeachingMethod.String())).
			WithModificationTime(time.Now(), time.Now()).
			WithSubjects(c.SubjectIds)
		course, err := builder.Build(ctx, m.DB)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		for _, locationID := range c.LocationIds {
			builder := domain.NewCourseAccessPath().
				WithCourseID(course.CourseID).
				WithLocationID(locationID).
				WithModificationTime(time.Now(), time.Now()).
				WithID(idutil.ULIDNow())
			if courseAP, err := builder.Build(ctx); err == nil {
				courseAccessPaths = append(courseAccessPaths, courseAP)
			}
		}
		domainCourses = append(domainCourses, course)
		courseIDs = append(courseIDs, course.CourseID)
	}
	courseLocationsActivePayload := queries.GetLocationsBelongToActiveStudentSubscriptionsByCourses{CourseIDs: courseIDs}
	courseLocationsActive, err := m.StudentSubscriptionCommandHandler.GetLocationsBelongToActiveStudentSubscriptionsByCourses(ctx, courseLocationsActivePayload)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("GetLocationsBelongToActiveStudentSubscriptionsByCourses err: %w", err).Error())
	}
	checkValid := isRequestValidLocations(courseLocationsActive, req.Courses)
	if !checkValid {
		return nil, status.Error(codes.AlreadyExists, fmt.Errorf("ra.manabie-error.already_exists").Error())
	}
	if err = m.CourseCommandHandler.UpsertCourses(ctx, commands.UpdateCoursesCommand{
		Courses:           domainCourses,
		CourseAccessPaths: courseAccessPaths,
		CourseIDs:         courseIDs,
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("CourseCommandHandler.UpsertCourses: %w", err).Error())
	}
	return &mpb.UpsertCoursesResponse{
		Successful: true,
	}, nil
}

func (m *MasterDataCourseService) GetCoursesByIDs(ctx context.Context, req *mpb.GetCoursesByIDsRequest) (*mpb.GetCoursesByIDsResponse, error) {
	courseIDs := req.CourseIds
	courses, err := m.CourseQueryHandler.GetCoursesByIDs(ctx, queries.GetCoursesByIDsPayload{IDs: courseIDs})
	if err != nil {
		return &mpb.GetCoursesByIDsResponse{}, status.Error(codes.Internal, err.Error())
	}
	pc := make([]*mpb.Course, len(courses))
	for i, v := range courses {
		pc[i] = &mpb.Course{
			Id:           v.CourseID,
			Name:         v.Name,
			CourseTypeId: v.CourseTypeID,
		}
	}
	return &mpb.GetCoursesByIDsResponse{Courses: pc}, nil
}

func isRequestValidLocations(courseLocationsActive map[string][]string, request []*mpb.UpsertCoursesRequest_Course) bool {
	for _, course := range request {
		if locations, ok := courseLocationsActive[course.Id]; ok {
			for _, locationID := range locations {
				if !golibs.InArrayString(locationID, course.LocationIds) {
					return false
				}
			}
		}
	}
	return true
}

func mapCourseCSVtoCourse(schoolID int64) func(c *validators.CSVLineValue[domain.Course]) *domain.Course {
	return func(c *validators.CSVLineValue[domain.Course]) *domain.Course {
		var id string
		if c.Value.CourseID == "" {
			id = idutil.ULIDNow()
		} else {
			id = c.Value.CourseID
		}
		return &domain.Course{
			CourseID:       id,
			CourseTypeID:   c.Value.CourseTypeID,
			Name:           c.Value.Name,
			IsArchived:     c.Value.IsArchived,
			PartnerID:      c.Value.PartnerID,
			TeachingMethod: c.Value.TeachingMethod,
			Remarks:        c.Value.Remarks,
			SchoolID:       int(schoolID),
		}
	}
}

func transformCSVLineToCourse(s []string) (*domain.Course, error) {
	course := &domain.Course{}
	const (
		CourseID = iota
		CourseName
		CourseTypeID
		CoursePartnerID
		Remarks
	)
	course.CourseID = s[CourseID]

	name := s[CourseName]
	if len(name) < 1 {
		return course, fmt.Errorf("%s", "name can not be empty")
	}
	if !utf8.ValidString(s[CourseName]) {
		return course, fmt.Errorf("%s", "name is not a valid UTF8 string")
	}
	course.Name = name

	typeID := s[CourseTypeID]
	course.CourseTypeID = typeID

	course.PartnerID = s[CoursePartnerID]
	course.Remarks = s[Remarks]

	return course, nil
}

func transformCSVLineToCourseV2(s []string) (*domain.Course, error) {
	course := &domain.Course{}
	const (
		CourseID = iota
		CourseName
		CourseTypeID
		CoursePartnerID
		TeachingMethod
		Remarks
	)
	course.CourseID = s[CourseID]

	name := s[CourseName]
	if len(name) < 1 {
		return course, fmt.Errorf("%s", "name can not be empty")
	}
	if !utf8.ValidString(s[CourseName]) {
		return course, fmt.Errorf("%s", "name is not a valid UTF8 string")
	}
	course.Name = name

	typeID := s[CourseTypeID]
	course.CourseTypeID = typeID

	course.PartnerID = s[CoursePartnerID]
	teachingMethod := s[TeachingMethod]
	if len(teachingMethod) > 0 && !slices.Contains([]string{"Group", "Individual"}, teachingMethod) {
		return course, fmt.Errorf("%s", "teachingMethod must be group or individual")
	}
	course.TeachingMethod = convertStringToTeachingMethod(teachingMethod)
	course.Remarks = s[Remarks]

	return course, nil
}

func convertStringToTeachingMethod(teachingMethod string) domain.CourseTeachingMethod {
	teachingMethods := map[string]domain.CourseTeachingMethod{
		"Group":      domain.CourseTeachingMethodGroup,
		"Individual": domain.CourseTeachingMethodIndividual,
	}

	return teachingMethods[teachingMethod]
}

func mapToCourseTypeIDs(c *validators.CSVLineValue[domain.Course]) string {
	return c.Value.CourseTypeID
}

func compareByCourseTypeIDs(c1, c2 *domain.CourseType) bool {
	return c1.CourseTypeID == c2.CourseTypeID
}
