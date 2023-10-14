package controller

import (
	"context"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	course_infras "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseTypeService struct {
	DB                       database.Ext
	CourseTypeRepo           course_infras.CourseTypeRepo
	CourseTypeCommandHandler commands.CourseTypeCommandHandler
}

func (c *CourseTypeService) ImportCourseTypes(ctx context.Context, req *mpb.ImportCourseTypesRequest) (res *mpb.ImportCourseTypesResponse, err error) {
	config := validators.CSVImportConfig[domain.CourseType]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "course_type_id",
				Required: false,
			},
			{
				Column:   "course_type_name",
				Required: true,
			},
			{
				Column:   "is_archived",
				Required: true,
			},
			{
				Column:   "remarks",
				Required: false,
			},
		},
		Transform: transformCSVLineToCourseType,
	}
	csvCourseTypes, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvCourseTypes, validators.GetErrorFromCSVValue[domain.CourseType], validators.HasCSVErr[domain.CourseType])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	courseTypes := sliceutils.Map(csvCourseTypes, mapCourseCSVtoCourseType)
	payload := commands.ImportCourseTypesPayload{
		CourseTypes: courseTypes,
	}
	err = c.CourseTypeCommandHandler.ImportCourseTypes(ctx, payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportCourseTypesResponse{}, nil
}

func mapCourseCSVtoCourseType(c *validators.CSVLineValue[domain.CourseType]) *domain.CourseType {
	var id string
	if c.Value.CourseTypeID == "" {
		id = idutil.ULIDNow()
	} else {
		id = c.Value.CourseTypeID
	}
	return &domain.CourseType{
		CourseTypeID: id,
		Name:         c.Value.Name,
		IsArchived:   c.Value.IsArchived,
		Remarks:      c.Value.Remarks,
	}
}

func transformCSVLineToCourseType(s []string) (*domain.CourseType, error) {
	ct := &domain.CourseType{}
	const (
		CourseTypeID = iota
		CourseTypeName
		IsArchived
		Remarks
	)

	typeID := s[CourseTypeID]
	if len(typeID) < 1 {
		typeID = idutil.ULIDNow()
	}
	ct.CourseTypeID = typeID

	name := s[CourseTypeName]
	if len(name) < 1 {
		return ct, fmt.Errorf("%s", "name can not be empty")
	}
	if !utf8.ValidString(s[CourseTypeName]) {
		return ct, fmt.Errorf("%s", "name is not a valid UTF8 string")
	}
	ct.Name = name

	isArchived, err := strconv.ParseBool(s[IsArchived])
	if err != nil {
		return ct, fmt.Errorf("%s is not a valid boolean: %s", s[IsArchived], err.Error())
	}
	ct.IsArchived = isArchived

	ct.Remarks = s[Remarks]

	return ct, nil
}
