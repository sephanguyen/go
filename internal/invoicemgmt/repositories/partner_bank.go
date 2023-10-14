package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type PartnerBankRepo struct {
}

func (r *PartnerBankRepo) RetrievePartnerBankByID(ctx context.Context, db database.QueryExecer, partnerBankID string) (*entities.PartnerBank, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerBankRepo.RetrievePartnerBankByID")
	defer span.End()

	e := &entities.PartnerBank{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE partner_bank_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &partnerBankID).ScanOne(e)

	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *PartnerBankRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.PartnerBank) error {
	ctx, span := interceptors.StartSpan(ctx, "PartnerBankRepo.Upsert")
	defer span.End()

	fields, values := e.FieldMapForUpsert()
	placeHolders := database.GeneratePlaceholders(len(fields))
	if e.PartnerBankID.String == "" {
		e.PartnerBankID = database.Text(idutil.ULIDNow())
	}
	stmt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (partner_bank_id) DO UPDATE SET %s;",
		e.TableName(),
		strings.Join(fields, ","),
		placeHolders,
		e.UpdateOnConflictQuery(),
	)

	cmd, err := db.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("err upsert partner bank: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("err upsert partner bank: %d RowsAffected", cmd.RowsAffected())
	}

	return nil
}

func (r *PartnerBankRepo) FindOne(ctx context.Context, db database.QueryExecer) (*entities.PartnerBank, error) {
	ctx, span := interceptors.StartSpan(ctx, "PartnerBankRepo.FindOne")
	defer span.End()

	e := &entities.PartnerBank{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE is_archived = false AND deleted_at IS null AND is_default = true ORDER BY created_at DESC LIMIT 1",
		strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}
