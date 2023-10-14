package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"

	"github.com/jackc/pgtype"
)

type BookHierarchyFlatten struct {
	BookID             pgtype.Text
	ChapterID          pgtype.Text
	TopicID            pgtype.Text
	LearningMaterialID pgtype.Text
}

func (bHierarchy *BookHierarchyFlatten) ToEntity() domain.BookHierarchyFlatten {
	return domain.BookHierarchyFlatten{
		BookID:             bHierarchy.BookID.String,
		ChapterID:          bHierarchy.ChapterID.String,
		TopicID:            bHierarchy.TopicID.String,
		LearningMaterialID: bHierarchy.LearningMaterialID.String,
	}
}
