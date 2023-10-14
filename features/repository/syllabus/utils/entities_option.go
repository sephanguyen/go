package utils

import "github.com/manabie-com/backend/internal/eureka/entities"

type StudyPlanItemOption func(u *entities.StudyPlanItem) error
type StudyPlanOption func(u *entities.StudyPlan) error
