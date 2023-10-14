package controller

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	locationInfras "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseAccessPathService struct {
	CourseAccessPathCommandHandler commands.CourseAccessPathCommandHandler
	CourseAccessPathQueryHandler   queries.CourseAccessPathQueryHandler
}

func NewCourseAccessPathService(db database.Ext,
	courseAccessPathRepo infrastructure.CourseAccessPathRepo,
	locationRepo locationInfras.LocationRepo,
	courseRepo infrastructure.CourseRepo) *CourseAccessPathService {
	return &CourseAccessPathService{
		CourseAccessPathCommandHandler: commands.CourseAccessPathCommandHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
			LocationRepo:         locationRepo,
			CourseRepo:           courseRepo,
		},
		CourseAccessPathQueryHandler: queries.CourseAccessPathQueryHandler{
			DB:                   db,
			CourseAccessPathRepo: courseAccessPathRepo,
		},
	}
}

func (c *CourseAccessPathService) ImportCourseAccessPaths(ctx context.Context, req *mpb.ImportCourseAccessPathsRequest) (res *mpb.ImportCourseAccessPathsResponse, err error) {
	config := validators.CSVImportConfig[domain.CourseAccessPath]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "course_access_path_id",
				Required: false,
			},
			{
				Column:   "course_id",
				Required: true,
			},
			{
				Column:   "location_id",
				Required: true,
			},
		},
		Transform: transformCSVLineToCAP,
	}
	csvCAP, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	rowErrors := sliceutils.MapSkip(csvCAP, validators.GetErrorFromCSVValue[domain.CourseAccessPath], validators.HasCSVErr[domain.CourseAccessPath])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	// validate existence of location id and course id
	_ = c.CourseAccessPathCommandHandler.CheckCoursesAndLocations(ctx, csvCAP)

	rowErrors = sliceutils.MapSkip(csvCAP, validators.GetErrorFromCSVValue[domain.CourseAccessPath], validators.HasCSVErr[domain.CourseAccessPath])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	courses := sliceutils.Map(csvCAP, mapCsvCAPtoCAP)
	payload := commands.UpsertCourseAccessPathsCommand{
		CourseAccessPaths: courses,
	}
	err = c.CourseAccessPathCommandHandler.UpsertCourseAccessPaths(ctx, payload)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mpb.ImportCourseAccessPathsResponse{}, nil
}

func (c *CourseAccessPathService) ExportCourseAccessPaths(ctx context.Context, req *mpb.ExportCourseAccessPathsRequest) (res *mpb.ExportCourseAccessPathsResponse, err error) {
	csv, err := c.CourseAccessPathQueryHandler.ExportCourseAccessPaths(ctx)
	if err != nil {
		return nil, err
	}
	return &mpb.ExportCourseAccessPathsResponse{
		Data: csv,
	}, nil
}

func transformCSVLineToCAP(s []string) (*domain.CourseAccessPath, error) {
	course := &domain.CourseAccessPath{}
	const (
		ID = iota
		CourseID
		LocationID
	)
	course.ID = s[ID]
	course.CourseID = s[CourseID]
	course.LocationID = s[LocationID]

	return course, nil
}

func mapCsvCAPtoCAP(c *validators.CSVLineValue[domain.CourseAccessPath]) *domain.CourseAccessPath {
	now := time.Now()
	id := c.Value.ID
	if id == "" {
		id = idutil.ULIDNow()
	}
	return &domain.CourseAccessPath{
		ID:         id,
		CourseID:   c.Value.CourseID,
		LocationID: c.Value.LocationID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
