package entity

import sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

// LearningMaterialPb combine LM proto struct + displayorder
type LearningMaterialPb struct {
	*sspb.LearningMaterialBase
	DisplayOrder  int32
	ParamPosition int // for validate swap.
}
