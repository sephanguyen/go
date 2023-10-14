package service

import (
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ExportService struct {
	DB database.Ext
}

func NewExportService(db database.Ext) (exportService *ExportService) {
	return &ExportService{
		DB: db,
	}
}
