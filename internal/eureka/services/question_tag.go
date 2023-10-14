package services

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QuestionTagService struct {
	sspb.UnimplementedQuestionTagServer
	DB database.Ext

	QuestionTagRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, e []*entities.QuestionTag) error
	}
}

func NewQuestionTagService(db database.Ext) sspb.QuestionTagServer {
	return &QuestionTagService{
		DB:              db,
		QuestionTagRepo: new(repositories.QuestionTagRepo),
	}
}

func validateCSVFormatQuestionTag(sc scanner.CSVScanner) error {
	if len(sc.GetRow()) == 0 {
		return fmt.Errorf("no data in csv file")
	}

	if len(sc.GetRow()) != 3 {
		return fmt.Errorf("csv file invalid format - number of column should be 3")
	}

	if sc.GetRow()[0] != "id" {
		return fmt.Errorf("csv file invalid format - first column (toLowerCase) should be 'id'")
	}
	if sc.GetRow()[1] != "name" {
		return fmt.Errorf("csv file invalid format - second column (toLowerCase) should be 'name'")
	}
	if sc.GetRow()[2] != "question_tag_type_id" {
		return fmt.Errorf("csv file invalid format - third column (toLowerCase) should be 'question_tag_type_id'")
	}
	return nil
}

func toQuestionTagEnt(id string, name string, questionTagTypeId string) (*entities.QuestionTag, error) {
	e := &entities.QuestionTag{}
	database.AllNullEntity(e)
	now := time.Now()
	if id == "" {
		id = idutil.ULIDNow()
	}
	if err := multierr.Combine(
		e.QuestionTagID.Set(id),
		e.Name.Set(name),
		e.QuestionTagTypeID.Set(questionTagTypeId),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *QuestionTagService) ImportQuestionTag(ctx context.Context, req *sspb.ImportQuestionTagRequest) (*sspb.ImportQuestionTagResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(req.Payload))

	err := validateCSVFormatQuestionTag(sc)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateCSVFormatQuestionTag failed, err: %w", err).Error())
	}

	questionTags := []*entities.QuestionTag{}
	for sc.Scan() {
		id := sc.Text("id")
		name := sc.Text("name")
		questionTagTypeId := sc.Text("question_tag_type_id")

		line := sc.GetCurRow()
		if name == "" {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.ImportQuestionTag cannot convert to question tag entity, err: name cannot be empty, at line %d", line).Error())
		}
		if questionTagTypeId == "" {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.ImportQuestionTag cannot convert to question tag entity, err: question tag type id cannot be empty, at line %d", line).Error())
		}

		questionTag, err := toQuestionTagEnt(id, name, questionTagTypeId)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("s.ImportQuestionTag cannot convert to question tag entity, err: %w", err).Error())
		}
		questionTags = append(questionTags, questionTag)
	}

	if len(questionTags) == 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("s.ImportQuestionTag, err: no data in csv file").Error())
	}

	if err := checkDuplicatedQuestionTagIDs(questionTags); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("checkDuplicatedQuestionTagIDs: %w", err).Error())
	}

	if err := s.QuestionTagRepo.BulkUpsert(ctx, s.DB, questionTags); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.QuestionTagRepo.BulkUpsert, err: %w", err).Error())
	}
	return &sspb.ImportQuestionTagResponse{}, nil
}

func checkDuplicatedQuestionTagIDs(items []*entities.QuestionTag) error {
	itemMap := make(map[string]int)
	for idx, item := range items {
		if v, ok := itemMap[item.QuestionTagID.String]; ok {
			// index start = 0
			// csv QuestionTag rows start = 2
			first := v + 2
			second := idx + 2
			return fmt.Errorf("duplicated id: %s at line %d and %d", item.QuestionTagID.String, first, second)
		}
		itemMap[item.QuestionTagID.String] = idx
	}
	return nil
}
