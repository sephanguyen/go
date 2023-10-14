package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type PartnerFormConfigRepo struct{}

func (p *PartnerFormConfigRepo) FindByPartnerAndFeatureName(ctx context.Context, db database.Ext, partnerID pgtype.Int4, featureName pgtype.Text) (*entities.PartnerFormConfig, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerFormConfigRepo.FindByPartnerAndFeatureName")
	defer span.End()

	config := &entities.PartnerFormConfig{}
	fields, values := config.FieldMap()
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

	return config, nil
}

func (p *PartnerFormConfigRepo) FindByFeatureName(ctx context.Context, db database.Ext, featureName pgtype.Text) (*entities.PartnerFormConfig, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerFormConfigRepo.FindByFeatureName")
	defer span.End()

	config := &entities.PartnerFormConfig{}
	fields, values := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM partner_form_configs
		WHERE feature_name = $1
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &featureName).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return config, nil
}
