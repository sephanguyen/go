package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"

type ImportGradesPayload struct {
	Grades []*domain.Grade
}
