package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/application/services/form_partner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"

	"github.com/jackc/pgx/v4"
)

type UpdaterIndividualLessonReport struct {
	DB                     database.Ext
	LessonReportRepo       infrastructure.LessonReportRepo
	LessonReportDetailRepo infrastructure.LessonReportDetailRepo
	PartnerFormConfigRepo  infrastructure.PartnerFormConfigRepo
	LessonMemberRepo       lesson_infrastructure.LessonMemberRepo
}

func (u *UpdaterIndividualLessonReport) Update(evictionPartner form_partner.EvictionPartner, ctx context.Context, lessonReport *domain.LessonReport) error {
	lessonReportDetails, err := u.LessonReportDetailRepo.GetDetailByLessonReportID(ctx, u.DB, lessonReport.LessonReportID)
	if err != nil {
		return err
	}
	lessonMembers := make([]*lesson_domain.UpdateLessonMemberReport, 0, len(lessonReportDetails))
	newFields := make([]*domain.PartnerDynamicFormFieldValue, 0, len(lessonReportDetails))
	converterFormReport := &form_partner.ConverterFormReport{
		EvictionPartner: evictionPartner,
	}
	for _, lessonReportDetail := range lessonReportDetails {
		fields, err := u.PartnerFormConfigRepo.GetMapStudentFieldValuesByDetailID(ctx, u.DB, lessonReportDetail.LessonReportDetailID)
		if err != nil {
			return err
		}
		studentId := lessonReportDetail.StudentID
		fieldsOfStudent, ok := fields[studentId]
		if !ok {
			fieldsOfStudent = make(domain.LessonReportFields, 0)
		}

		data, err := converterFormReport.Convert(lessonReport, lessonReportDetail, fieldsOfStudent)
		if err != nil {
			return err
		}
		lessonMembers = append(lessonMembers, data.LessonMemberUpdate)
		newFields = append(newFields, data.Fields...)
	}
	lessonReport.FormConfigID = evictionPartner.GetConfigFormId()
	return database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		if _, err := u.LessonReportRepo.Update(ctx, tx, lessonReport); err != nil {
			return fmt.Errorf("Update lesson report: %s", err)
		}

		if err := u.LessonReportDetailRepo.UpsertFieldValues(ctx, tx, newFields); err != nil {
			return fmt.Errorf("Update field values report: %s", err)
		}

		if err := u.LessonMemberRepo.UpdateLessonMembers(ctx, tx, lessonMembers); err != nil {
			return fmt.Errorf("Update lesson member: %s", err)
		}
		return nil
	})
}
