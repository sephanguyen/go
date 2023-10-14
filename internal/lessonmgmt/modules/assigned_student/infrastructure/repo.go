package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	masterdata_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
)

type AssignedStudentRepo interface {
	GetAssignedStudentList(ctx context.Context, db database.QueryExecer, params *payloads.GetAssignedStudentListArg) (asgStudents []*domain.AssignedStudent, total uint32, offsetID string, preTotal uint32, err error)
	GetStudentAttendance(ctx context.Context, db database.QueryExecer, filter domain.GetStudentAttendanceParams) ([]*domain.StudentAttendance, uint32, error)
}

type AcademicYearRepo interface {
	GetCurrentAcademicYear(ctx context.Context, db database.Ext) (*masterdata_domain.AcademicYear, error)
}
