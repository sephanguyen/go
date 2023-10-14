package entity

import (
	"github.com/manabie-com/backend/internal/eureka/entities"
)

type getStudyPlanItemsAble interface {
	getStudyPlanItems() []*entities.StudyPlanItem
}

type getStudyPlanAble interface {
	getStudyPlan() *entities.StudyPlan
}

func GetStudyPlanItems(entity getStudyPlanItemsAble) []*entities.StudyPlanItem {
	return entity.getStudyPlanItems()
}

func GetStudyPlan(entity getStudyPlanAble) *entities.StudyPlan {
	return entity.getStudyPlan()
}
