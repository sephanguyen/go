package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GradeService struct {
	ImportGradesCommandHandler commands.ImportGradesCommandHandler
	ExportGradesQueryHandler   queries.ExportGradesQueryHandler
}

func NewGradeService(
	db database.Ext,
	gradeRepo infrastructure.GradeRepo,
) *GradeService {
	return &GradeService{
		ImportGradesCommandHandler: commands.ImportGradesCommandHandler{
			DB:        db,
			GradeRepo: gradeRepo,
		},
		ExportGradesQueryHandler: queries.ExportGradesQueryHandler{
			DB:        db,
			GradeRepo: gradeRepo,
		},
	}
}

func (g *GradeService) ImportGrades(ctx context.Context, req *mpb.ImportGradesRequest) (res *mpb.ImportGradesResponse, err error) {
	config := validators.CSVImportConfig[domain.Grade]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "grade_id",
				Required: false,
			},
			{
				Column:   "grade_partner_id",
				Required: true,
			},
			{
				Column:   "name",
				Required: true,
			},
			{
				Column:   "sequence",
				Required: true,
			},
			{
				Column:   "remarks",
				Required: false,
			},
		},
		Transform: transformCSVLineToGrade,
	}
	csvGrades, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	csvGrades, _ = checkUniqueGrade(csvGrades)

	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvGrades, validators.GetErrorFromCSVValue[domain.Grade], validators.HasCSVErr[domain.Grade])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	grades := sliceutils.Map(csvGrades, mapGradeCSVtoGrade)
	payload := commands.ImportGradesPayload{
		Grades: grades,
	}

	err = g.ImportGradesCommandHandler.ImportGrades(ctx, payload)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportGradesResponse{}, nil
}

func (g *GradeService) ExportGrades(ctx context.Context, req *mpb.ExportGradesRequest) (res *mpb.ExportGradesResponse, err error) {
	bytes, err := g.ExportGradesQueryHandler.ExportGrades(ctx)
	if err != nil {
		return &mpb.ExportGradesResponse{}, err
	}
	res = &mpb.ExportGradesResponse{
		Data: bytes,
	}
	return res, nil
}

func transformCSVLineToGrade(s []string) (*domain.Grade, error) {
	g := &domain.Grade{}
	const (
		ID = iota
		PartnerID
		Name
		Sequence
		Remarks
	)
	errs := []error{}

	gID := s[ID]
	if len(gID) < 1 {
		gID = idutil.ULIDNow()
	}
	g.ID = gID

	g.PartnerInternalID = s[PartnerID]

	name := s[Name]
	if len(name) < 1 {
		errs = append(errs, fmt.Errorf("%s", "grade name can not be empty"))
	}
	if !utf8.ValidString(s[Name]) {
		errs = append(errs, fmt.Errorf("%s", "grade name is not a valid UTF8 string"))
	}
	g.Name = name

	sq, err := strconv.Atoi(strings.TrimSpace(s[Sequence]))
	if err != nil {
		errs = append(errs, fmt.Errorf("sequence is not a number: %s", s[Sequence]))
	}
	g.Sequence = sq
	g.Remarks = s[Remarks]
	if len(errs) > 0 {
		return g, errs[0]
	}
	return g, nil
}

func checkUniqueGrade(grades []*validators.CSVLineValue[domain.Grade]) ([]*validators.CSVLineValue[domain.Grade], bool) {
	sequenceMap := make(map[string]*validators.CSVLineValue[domain.Grade], len(grades))
	partnerIDMap := make(map[string]*validators.CSVLineValue[domain.Grade], len(grades))
	hasDuplication := false

	for i, g := range grades {
		v, ok := sequenceMap[fmt.Sprintf("%d", g.Value.Sequence)]
		if ok {
			if g.Error == nil {
				g.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("sequence %d is duplicated", v.Value.Sequence),
				}
				hasDuplication = true
			}
		} else {
			sequenceMap[fmt.Sprintf("%d", g.Value.Sequence)] = g
		}
		v, ok = partnerIDMap[g.Value.PartnerInternalID]
		if ok {
			if g.Error == nil {
				g.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("grade partner id %s is duplicated", v.Value.PartnerInternalID),
				}
				hasDuplication = true
			}
		} else {
			partnerIDMap[g.Value.PartnerInternalID] = g
		}
	}
	return grades, hasDuplication
}

func mapGradeCSVtoGrade(c *validators.CSVLineValue[domain.Grade]) *domain.Grade {
	var id string
	if c.Value.ID == "" {
		id = idutil.ULIDNow()
	} else {
		id = c.Value.ID
	}
	return &domain.Grade{
		ID:                id,
		PartnerInternalID: c.Value.PartnerInternalID,
		Name:              c.Value.Name,
		Sequence:          c.Value.Sequence,
		IsArchived:        c.Value.IsArchived,
		Remarks:           c.Value.Remarks,
	}
}
