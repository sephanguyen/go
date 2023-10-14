package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QuestionTagTypeService struct {
	sspb.QuestionTagTypeServer
	DB database.Ext

	QuestionTagTypeRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, data []*entities.QuestionTagType) error
	}
}

type QuestionTagTypeRow struct {
	ID   string
	Name string
}

func (r *QuestionTagTypeRow) toQuestionTagTypeEntity() *entities.QuestionTagType {
	e := &entities.QuestionTagType{
		QuestionTagTypeID: database.Text(r.ID),
		Name:              database.Text(r.Name),
	}
	e.Now()
	return e
}

func NewQuestionTagTypeService(db database.Ext) *QuestionTagTypeService {
	return &QuestionTagTypeService{
		DB:                  db,
		QuestionTagTypeRepo: &repositories.QuestionTagTypeRepo{},
	}
}

func (q *QuestionTagTypeService) ImportQuestionTagTypes(ctx context.Context, req *sspb.ImportQuestionTagTypesRequest) (*sspb.ImportQuestionTagTypesResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(req.Payload))

	err := validateCSVFormat(sc)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateCSVFormat: %w", err).Error())
	}

	tagTypes := []*entities.QuestionTagType{}
	for sc.Scan() {
		id := sc.Text("id")
		name := sc.Text("name")
		row, err := newQuestionTagTypeRow(id, name)
		if err != nil {
			line := sc.GetCurRow()
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("newQuestionTagTypeRow: %w at line %d", err, line).Error())
		}
		tagType := row.toQuestionTagTypeEntity()
		tagTypes = append(tagTypes, tagType)
	}
	// if no rows
	if len(tagTypes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no data in csv file")
	}

	err = checkDuplicatedQuestionTagTypeIDs(tagTypes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("checkDuplicatedQuestionTagTypeIDs: %w", err).Error())
	}
	err = q.QuestionTagTypeRepo.BulkUpsert(ctx, q.DB, tagTypes)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.ImportQuestionTagTypesResponse{}, nil
}

func validateCSVFormat(sc scanner.CSVScanner) error {
	// no columns
	if len(sc.GetRow()) == 0 {
		return fmt.Errorf("no data in csv file")
	}
	// length columns
	if len(sc.GetRow()) != 2 {
		return fmt.Errorf("csv file has invalid format - number of column should be 2")
	}
	if sc.GetRow()[0] != "id" {
		return fmt.Errorf("csv file has invalid format - first column (toLowerCase) should be 'id'")
	}
	if sc.GetRow()[1] != "name" {
		return fmt.Errorf("csv file has invalid format - second column (toLowerCase) should be 'name'")
	}
	return nil
}

func newQuestionTagTypeRow(id, name string) (*QuestionTagTypeRow, error) {
	if id == "" {
		id = idutil.ULIDNow()
	}
	if name == "" {
		return nil, fmt.Errorf("name be empty!")
	}
	return &QuestionTagTypeRow{
		ID:   id,
		Name: name,
	}, nil
}

func checkDuplicatedQuestionTagTypeIDs(items []*entities.QuestionTagType) error {
	itemMap := make(map[string]int)
	for idx, item := range items {
		if v, ok := itemMap[item.QuestionTagTypeID.String]; ok {
			// index start = 0
			// csv QuestionTagType rows start = 2
			first := v + 2
			second := idx + 2
			return fmt.Errorf("duplicated id: %s at line %d and %d", item.QuestionTagTypeID.String, first, second)
		}
		itemMap[item.QuestionTagTypeID.String] = idx
	}
	return nil
}
