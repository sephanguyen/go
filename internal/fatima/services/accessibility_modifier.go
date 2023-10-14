package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
)

type AccessibilityModifyService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	StudentPackageRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudentPackage) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) error
		SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
	}

	StudentPackageAccessPathRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, ents []*entities.StudentPackageAccessPath) error
		DeleteByStudentPackageIDs(ctx context.Context, db database.QueryExecer, spIDs pgtype.TextArray) error
	}
}

// SyncStudentPackage handle EventSyncStudentPackage event, upsert StudentPackage if ActionKind=UPSERTED and softDelete if ActionKind=DELETED.
func (c *AccessibilityModifyService) SyncStudentPackage(ctx context.Context, req *npb.EventSyncStudentPackage) error {
	return database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, request := range req.StudentPackages {
			switch request.ActionKind {
			case npb.ActionKind_ACTION_KIND_UPSERTED:
				studentPackages, err := toStudentPackage(request)
				if err != nil {
					return err
				}

				if err := c.StudentPackageRepo.SoftDelete(ctx, tx, database.Text(request.StudentId)); err != nil {
					return fmt.Errorf("err c.StudentPackageRepo.SoftDelete: %w", err)
				}

				if err := c.StudentPackageRepo.BulkInsert(ctx, tx, studentPackages); err != nil {
					return fmt.Errorf("err s.StudentPackageRepo.BulkInsert, studentID %s: %w",
						request.StudentId,
						err)
				}
				for _, sp := range studentPackages {
					spAccessPaths, err := generateListStudentPackageAccessPathsFromStudentPackage(sp)
					if err != nil {
						return err
					}
					if err = c.StudentPackageAccessPathRepo.BulkUpsert(ctx, tx, spAccessPaths); err != nil {
						return err
					}
				}
			case npb.ActionKind_ACTION_KIND_DELETED:
				studentPackage := request
				ids := []string{}
				for index := 0; index < len(studentPackage.Packages); index++ {
					value := studentPackage.Packages[index]
					ids = append(ids, getStudentPackageID(request.StudentId, value))
				}

				if err := c.StudentPackageRepo.SoftDeleteByIDs(ctx, tx, database.TextArray(ids)); err != nil {
					return fmt.Errorf("err c.StudentPackageRepo.SoftDelete: %w", err)
				}
			}
		}

		data, err := proto.Marshal(req)
		if err != nil {
			return fmt.Errorf("unable marshal: %w", err)
		}

		_, err = c.JSM.PublishAsyncContext(ctx, constants.SubjectSyncJprepStudentPackageEventNats, data)
		if err != nil {
			return fmt.Errorf("Fatima.SyncStudentPackage s.JSM.PublishAsyncContext failed: %v", err)
		}
		return nil
	})
}
