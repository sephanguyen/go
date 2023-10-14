package dto

import "github.com/jackc/pgtype"

type BookContent struct {
	ID   pgtype.Text
	Name pgtype.Text

	ChapterJSON pgtype.Text
}

type Chapter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`

	Topics []Topic `json:"topics"`
}

type Topic struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Order   int    `json:"order"`
	IconURL string `json:"iconUrl"`

	LearningMaterials []BookContentLearningMaterial `json:"materials"`
}

type BookContentLearningMaterial struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
	Type  string `json:"type"`
}
