package controller

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/dto"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/validators"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubjectService struct {
	ImportSubjectsCommandHandler commands.ImportSubjectsCommandHandler
	ExportSubjectsQueryHandler   queries.ExportSubjectsQueryHandler
}

func NewSubjectService(
	db database.Ext,
	subjectRepo infrastructure.SubjectRepo,
) *SubjectService {
	return &SubjectService{
		ImportSubjectsCommandHandler: commands.ImportSubjectsCommandHandler{
			DB:          db,
			SubjectRepo: subjectRepo,
		},
		ExportSubjectsQueryHandler: queries.ExportSubjectsQueryHandler{
			DB:          db,
			SubjectRepo: subjectRepo,
		},
	}
}

func (s *SubjectService) ImportSubjects(ctx context.Context, req *mpb.ImportSubjectsRequest) (res *mpb.ImportSubjectsResponse, err error) {
	config := validators.CSVImportConfig[domain.Subject]{
		ColumnConfig: []validators.CSVColumn{
			{
				Column:   "subject_id",
				Required: false,
			},
			{
				Column:   "name",
				Required: true,
			},
		},
		Transform: transformCSVLineToSubject,
	}
	csvSubjects, err := validators.ReadAndValidateCSV(req.Payload, config)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	csvSubjects, _ = checkUniqueSubject(csvSubjects)

	// collect error lines only
	rowErrors := sliceutils.MapSkip(csvSubjects, validators.GetErrorFromCSVValue[domain.Subject], validators.HasCSVErr[domain.Subject])

	if len(rowErrors) > 0 {
		return nil, utils.GetValidationError(rowErrors)
	}

	subjects := sliceutils.Map(csvSubjects, mapSubjectCSVtoSubject)
	payload := commands.ImportSubjectsPayload{
		Subjects: subjects,
	}

	err = s.ImportSubjectsCommandHandler.ImportSubjects(ctx, payload)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mpb.ImportSubjectsResponse{}, nil
}

func (s *SubjectService) ExportSubjects(ctx context.Context, _ *mpb.ExportSubjectsRequest) (res *mpb.ExportSubjectsResponse, err error) {
	bytes, err := s.ExportSubjectsQueryHandler.ExportSubjects(ctx)
	if err != nil {
		return &mpb.ExportSubjectsResponse{}, err
	}
	res = &mpb.ExportSubjectsResponse{
		Data: bytes,
	}
	return res, nil
}

func transformCSVLineToSubject(s []string) (*domain.Subject, error) {
	subject := &domain.Subject{}
	const (
		ID = iota
		Name
	)
	errs := []error{}

	sID := s[ID]
	if len(sID) < 1 {
		sID = idutil.ULIDNow()
	}
	subject.SubjectID = sID

	name := s[Name]
	if len(name) < 1 {
		errs = append(errs, fmt.Errorf("%s", "subject name can not be empty"))
	}
	if !utf8.ValidString(s[Name]) {
		errs = append(errs, fmt.Errorf("%s", "subject name is not a valid UTF8 string"))
	}
	subject.Name = name

	if len(errs) > 0 {
		return subject, errs[0]
	}
	return subject, nil
}

func mapSubjectCSVtoSubject(c *validators.CSVLineValue[domain.Subject]) *domain.Subject {
	return &domain.Subject{
		SubjectID: c.Value.SubjectID,
		Name:      c.Value.Name,
	}
}

func checkUniqueSubject(subjects []*validators.CSVLineValue[domain.Subject]) ([]*validators.CSVLineValue[domain.Subject], bool) {
	idMap := make(map[string]*validators.CSVLineValue[domain.Subject], len(subjects))
	hasDuplication := false

	for i, s := range subjects {
		v, ok := idMap[s.Value.SubjectID]
		if ok {
			if s.Error == nil {
				s.Error = &dto.UpsertError{
					RowNumber: int32(i + 2),
					Error:     fmt.Sprintf("id %s is duplicated", v.Value.SubjectID),
				}
				hasDuplication = true
			}
		} else {
			idMap[s.Value.SubjectID] = s
		}
	}
	return subjects, hasDuplication
}
