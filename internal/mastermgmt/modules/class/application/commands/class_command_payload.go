package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

type CreateClass struct {
	Classes []*domain.Class
}

type UpdateClassById struct {
	ID   string
	Name string
}

type DeleteClassById struct {
	ID string
}
