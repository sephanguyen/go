package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
)

type PartnerFormConfigRepo struct{}

func (p *PartnerFormConfigRepo) FindByPartnerAndFeatureName(ctx context.Context, db database.Ext, partnerID int, featureName string) (*domain.PartnerFormConfig, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerFormConfigRepo.FindByPartnerAndFeatureName")
	defer span.End()

	configDTO := &PartnerFormConfigDTO{}
	fields, values := configDTO.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM partner_form_configs
		WHERE partner_id = $1
			AND feature_name = $2
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &partnerID, &featureName).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	// create domain
	domain, err := domain.NewPartnerFormConfigBuilder().
		WithFormConfigID(configDTO.FormConfigID.String).
		WithFeatureName(configDTO.FeatureName.String).
		WithFormConfigData(configDTO.FormConfigData.Bytes).
		WithModificationTime(configDTO.CreatedAt.Time, configDTO.UpdatedAt.Time).
		WithPartnerID(int(configDTO.PartnerID.Int)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("PartnerFormConfigRepo.FindByPartnerAndFeatureName(): Error parsing DTO to domain: %w", err)
	}
	return domain, nil
}

func (p *PartnerFormConfigRepo) DeleteByLessonReportDetailIDs(ctx context.Context, db database.Ext, ids []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.DeleteByLessonReportDetailIDs")
	defer span.End()
	dto := PartnerDynamicFormFieldValueDTO{}
	query := fmt.Sprintf(`UPDATE %s SET updated_at = now(), deleted_at = now()
	 WHERE lesson_report_detail_id = ANY($1) AND deleted_at IS NULL`, dto.TableName())

	_, err := db.Exec(ctx, query, database.TextArray(ids))

	if err != nil {
		return fmt.Errorf("PartnerFormConfigRepo.DeleteByLessonReportDetailIDs:%w", err)
	}
	return nil
}

func (p *PartnerFormConfigRepo) GetMapStudentFieldValuesByDetailID(ctx context.Context, db database.Ext, lessonReportDetailId string) (map[string]domain.LessonReportFields, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportDetailRepo.GetMapStudentFieldValuesByDetailID")
	defer span.End()

	values := PartnerDynamicFormFieldValueWithStudentIdDTOs{}

	query := fmt.Sprintf(`SELECT pdffv.dynamic_form_field_value_id, pdffv.field_id, pdffv.lesson_report_detail_id,
			pdffv.created_at, pdffv.updated_at, pdffv.deleted_at, pdffv.value_type, pdffv.int_value, 
			pdffv.string_value, pdffv.bool_value, pdffv.string_array_value, pdffv.int_array_value, pdffv.string_set_value,
			pdffv.int_set_value, pdffv.field_render_guide, lrd.student_id
		FROM partner_dynamic_form_field_values pdffv  
		left join lesson_report_details lrd on pdffv.lesson_report_detail_id  = lrd.lesson_report_detail_id
		Where lrd.lesson_report_detail_id = $1 AND pdffv.deleted_at IS NULL AND lrd.deleted_at IS NULL
	`)

	err := database.Select(ctx, db, query, &lessonReportDetailId).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	mapFieldValuesOfStudent := make(map[string]domain.LessonReportFields)
	for _, v := range values {
		key := v.StudentID.String
		fieldValues, ok := mapFieldValuesOfStudent[key]
		if !ok {
			fieldValues = make(domain.LessonReportFields, 0, len(values))
		}
		fieldValues = append(fieldValues, v.ToDomain())
		mapFieldValuesOfStudent[key] = fieldValues
	}

	return mapFieldValuesOfStudent, nil
}

func (p *PartnerFormConfigRepo) CreatePartnerFormConfig(ctx context.Context, db database.Ext, partnerFormConfig *domain.PartnerFormConfig) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerFormConfigRepo.CreatePartnerFormConfig")
	defer span.End()

	configDTO, err := NewPartnerFormConfigFromEntity(partnerFormConfig)
	if err != nil {
		return err
	}

	fieldNames, args := configDTO.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO partner_form_configs (%s) VALUES (%s) 
	ON CONFLICT  ON CONSTRAINT partner_form_configs_pk 
	DO NOTHING`,
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
