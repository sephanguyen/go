package entities

import (
	"github.com/manabie-com/backend/internal/bob/entities"

	"github.com/manabie-com/backend/internal/yasuo/types"
)

type School struct {
	entities.School
	Point types.Point `sql:"type:geometry"`
}

type SchoolExpand struct {
	TableName struct{} `sql:"schools"`
	School
	Latitude  string
	Longitude string
}
