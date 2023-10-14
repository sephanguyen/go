package entities

import (
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuestionHierarchyType string

const (
	QuestionHierarchyQuestion      QuestionHierarchyType = "QUESTION"
	QuestionHierarchyQuestionGroup QuestionHierarchyType = "QUESTION_GROUP"
)

type QuestionGroup struct {
	BaseEntity
	QuestionGroupID    pgtype.Text  `sql:"question_group_id"`
	LearningMaterialID pgtype.Text  `sql:"learning_material_id"`
	Name               pgtype.Text  `sql:"name"`
	Description        pgtype.Text  `sql:"description"`
	RichDescription    pgtype.JSONB `sql:"rich_description"`

	totalChildren pgtype.Int4
	totalPoints   pgtype.Int4
}

func (q *QuestionGroup) TotalChildren() *pgtype.Int4 {
	return &q.totalChildren
}

func (q *QuestionGroup) TotalPoints() *pgtype.Int4 {
	return &q.totalPoints
}

func (q *QuestionGroup) SetTotalChildrenAndPoints(totalChildren, totalPoints int32) {
	q.totalChildren = database.Int4(totalChildren)
	q.totalPoints = database.Int4(totalPoints)
}

// FieldMap return a map of field name and pointer to field
func (q *QuestionGroup) FieldMap() ([]string, []interface{}) {
	return []string{
			"question_group_id",
			"learning_material_id",
			"name",
			"description",
			"rich_description",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&q.QuestionGroupID,
			&q.LearningMaterialID,
			&q.Name,
			&q.Description,
			&q.RichDescription,
			&q.CreatedAt,
			&q.UpdatedAt,
			&q.DeletedAt,
		}
}

func (q *QuestionGroup) FieldMapUpsert() ([]string, []interface{}) {
	return []string{
			"question_group_id",
			"learning_material_id",
			"name",
			"description",
			"rich_description",
		}, []interface{}{
			&q.QuestionGroupID,
			&q.LearningMaterialID,
			&q.Name,
			&q.Description,
			&q.RichDescription,
		}
}

// TableName returns "question_group" table name
func (q *QuestionGroup) TableName() string {
	return "question_group"
}

func (q *QuestionGroup) IsValidToUpsert() error {
	if q.LearningMaterialID.Status != pgtype.Present || len(q.LearningMaterialID.String) == 0 {
		return fmt.Errorf("LearningMaterialID could not null")
	}

	return nil
}

func (q *QuestionGroup) GetRichDescription() (*entities.RichText, error) {
	r := &entities.RichText{}
	err := q.RichDescription.AssignTo(r)
	return r, err
}

type QuestionGroups []*QuestionGroup

func (u *QuestionGroups) Add() database.Entity {
	e := &QuestionGroup{}
	*u = append(*u, e)
	return e
}

type QuestionHierarchyObj struct {
	ID          string                `json:"id"`
	Type        QuestionHierarchyType `json:"type"`
	ChildrenIDs []string              `json:"children_ids,omitempty"`
}

type QuestionHierarchy []*QuestionHierarchyObj

func (q *QuestionHierarchy) AddQuestionGroupID(ids ...string) {
	for _, id := range ids {
		*q = append(*q, &QuestionHierarchyObj{
			ID:   id,
			Type: QuestionHierarchyQuestionGroup,
		})
	}
}

func (q *QuestionHierarchy) AddQuestionID(ids ...string) {
	for _, id := range ids {
		*q = append(*q, &QuestionHierarchyObj{
			ID:   id,
			Type: QuestionHierarchyQuestion,
		})
	}
}

func (q *QuestionHierarchy) UnmarshalJSONBArray(jsonbArray pgtype.JSONBArray) error {
	for _, ele := range jsonbArray.Elements {
		var questionHierarchyObj *QuestionHierarchyObj
		if err := json.Unmarshal(ele.Bytes, &questionHierarchyObj); err != nil {
			return err
		}

		*q = append(*q, questionHierarchyObj)
	}
	return nil
}

func (q *QuestionHierarchy) CreateQuestionHierarchyMap() map[string]*QuestionHierarchyObj {
	questionHierarchyMap := make(map[string]*QuestionHierarchyObj)

	for _, obj := range *q {
		questionHierarchyMap[obj.ID] = obj
	}

	return questionHierarchyMap
}

func (q *QuestionHierarchy) CreateQuizExternalIDs() ([]string, error) {
	quizExternalIDs := []string{}

	for _, questionHierarchyObj := range *q {
		switch questionHierarchyObj.Type {
		case QuestionHierarchyQuestion:
			quizExternalIDs = append(quizExternalIDs, questionHierarchyObj.ID)
		case QuestionHierarchyQuestionGroup:
			quizExternalIDs = append(quizExternalIDs, questionHierarchyObj.ChildrenIDs...)
		default:
			return nil, fmt.Errorf("invalid question type")
		}
	}
	return quizExternalIDs, nil
}

func (q *QuestionHierarchy) AppendQuestionHierarchyFromPb(pbQuestionHierarchy []*sspb.QuestionHierarchy) {
	for _, questionHierarchyObj := range pbQuestionHierarchy {
		*q = append(*q, &QuestionHierarchyObj{
			ID:          questionHierarchyObj.Id,
			Type:        QuestionHierarchyType(questionHierarchyObj.Type.String()),
			ChildrenIDs: questionHierarchyObj.ChildrenIds,
		})
	}
}

func (q *QuestionHierarchy) IsIDDuplicated() error {
	questionHierarchyMap := make(map[string]bool)

	for _, questionHierarchyObj := range *q {
		if _, ok := questionHierarchyMap[questionHierarchyObj.ID]; ok {
			return fmt.Errorf("duplicate question id %s in new question hierarchy", questionHierarchyObj.ID)
		}

		questionHierarchyMap[questionHierarchyObj.ID] = true
	}
	return nil
}

func (q *QuestionHierarchy) IsElementsMatched(otherQuestionHierarchy QuestionHierarchy) error {
	otherQuestionHierarchyMap := otherQuestionHierarchy.CreateQuestionHierarchyMap()

	for _, questionHierarchyObj := range *q {
		if _, ok := otherQuestionHierarchyMap[questionHierarchyObj.ID]; !ok {
			return fmt.Errorf("question id %s not exist in current question hierarchy", questionHierarchyObj.ID)
		}

		if questionHierarchyObj.Type != otherQuestionHierarchyMap[questionHierarchyObj.ID].Type {
			return fmt.Errorf("question type mismatch, expected %s but got %s", otherQuestionHierarchyMap[questionHierarchyObj.ID].Type, questionHierarchyObj.Type)
		}

		if isChildrenIDsEqual := sliceutils.UnorderedEqual(questionHierarchyObj.ChildrenIDs, otherQuestionHierarchyMap[questionHierarchyObj.ID].ChildrenIDs); !isChildrenIDsEqual {
			return fmt.Errorf("mismatch children question id in question id %s", questionHierarchyObj.ID)
		}
	}
	return nil
}

func (q *QuestionHierarchy) ExcludeQuestionIDs(questionIDs []string) QuestionHierarchy {
	newQuestionHierarchy := QuestionHierarchy{}

	for _, questionHierarchyObj := range *q {
		isQuestionIDContained := sliceutils.Contains(questionIDs, questionHierarchyObj.ID)

		if isQuestionIDContained && questionHierarchyObj.Type == QuestionHierarchyQuestion {
			continue
		}

		if len(questionHierarchyObj.ChildrenIDs) != 0 {
			questionHierarchyObj.ChildrenIDs = stringutil.SliceElementsDiff(questionHierarchyObj.ChildrenIDs, questionIDs)
		}
		newQuestionHierarchy = append(newQuestionHierarchy, questionHierarchyObj)
	}
	return newQuestionHierarchy
}

func (q *QuestionHierarchy) ExcludeQuestionGroupIDs(questionGroupIDs []string) QuestionHierarchy {
	newQuestionHierarchy := QuestionHierarchy{}

	for _, questionHierarchyObj := range *q {
		isIDContained := sliceutils.Contains(questionGroupIDs, questionHierarchyObj.ID)

		if isIDContained && questionHierarchyObj.Type == QuestionHierarchyQuestionGroup {
			continue
		}

		newQuestionHierarchy = append(newQuestionHierarchy, questionHierarchyObj)
	}
	return newQuestionHierarchy
}

func (q *QuestionHierarchy) GetQuestionGroupIDs() []string {
	res := make([]string, 0)
	for _, item := range *q {
		if item.Type == QuestionHierarchyQuestionGroup {
			res = append(res, item.ID)
		}
	}

	return res
}

func (q *QuestionHierarchy) AddChildrenIDsForQuestionGroup(groupID string, childrenIDs ...string) error {
	for i := range *q {
		if (*q)[i].ID == groupID && (*q)[i].Type == QuestionHierarchyQuestionGroup {
			(*q)[i].ChildrenIDs = append((*q)[i].ChildrenIDs, childrenIDs...)
			return nil
		}
	}

	return fmt.Errorf("not found question group have id %s", groupID)
}

func QuestionGroupsToQuestionGroupProtoBufMess(q QuestionGroups) ([]*cpb.QuestionGroup, error) {
	res := make([]*cpb.QuestionGroup, 0, len(q))
	for _, element := range q {
		var richDescriptionEnt entities.RichText

		if err := element.RichDescription.AssignTo(&richDescriptionEnt); err != nil {
			return nil, err
		}

		richDescriptionPb := &cpb.RichText{
			Raw:      richDescriptionEnt.Raw,
			Rendered: richDescriptionEnt.RenderedURL,
		}

		res = append(res, &cpb.QuestionGroup{
			QuestionGroupId:    element.QuestionGroupID.String,
			LearningMaterialId: element.LearningMaterialID.String,
			Name:               element.Name.String,
			Description:        element.Description.String,
			RichDescription:    richDescriptionPb,
			CreatedAt:          timestamppb.New(element.CreatedAt.Time),
			UpdatedAt:          timestamppb.New(element.UpdatedAt.Time),
			TotalChildren:      element.totalChildren.Int,
			TotalPoints:        element.totalPoints.Int,
		})
	}

	return res, nil
}
