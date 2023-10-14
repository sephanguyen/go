package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type UsrEmailRepo struct{}

func (r *UsrEmailRepo) Create(ctx context.Context, db database.QueryExecer, usrID pgtype.Text, email pgtype.Text) (*entity.UsrEmail, error) {
	ctx, span := interceptors.StartSpan(ctx, "UsrEmailRepo.Create")
	defer span.End()

	now := time.Now()
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	email.String = strings.TrimSpace(email.String)
	email.String = strings.ToLower(email.String)

	usrEmailToCreate := &entity.UsrEmail{
		UsrID:        usrID,
		Email:        email,
		CreatedAt:    database.Timestamptz(now),
		UpdatedAt:    database.Timestamptz(now),
		DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
		ResourcePath: database.Text(resourcePath),
	}

	fieldsToCreate, valuesToCreate := database.GetFieldMapExcept(usrEmailToCreate, "import_id")

	createdUsrEmail := &entity.UsrEmail{}
	database.AllNullEntity(createdUsrEmail)
	createdFields, createdValues := createdUsrEmail.FieldMap()

	// Use DO UPDATE SET to guarantee the id will be
	// returned even the record is existed before
	stmt :=
		`
		INSERT INTO 
			%s(%s)
		VALUES 
			(%s) 
		ON CONFLICT 
			ON CONSTRAINT usr_email__pkey
		DO UPDATE SET 
			updated_at = EXCLUDED.updated_at 
		RETURNING 
			%s;
		`

	stmt = fmt.Sprintf(
		stmt,
		usrEmailToCreate.TableName(),
		strings.Join(fieldsToCreate, ","),
		database.GeneratePlaceholders(len(fieldsToCreate)),
		strings.Join(createdFields, ","),
	)

	err := db.QueryRow(ctx, stmt, valuesToCreate...).Scan(createdValues...)

	switch err {
	case nil:
		return createdUsrEmail, nil
	case pgx.ErrNoRows:
		return nil, errors.Wrap(err, "database.InsertReturning returns no row")
	default:
		return nil, errors.Wrap(err, "database.InsertReturning")
	}
}

func (r *UsrEmailRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser) ([]*entity.UsrEmail, error) {
	ctx, span := interceptors.StartSpan(ctx, "UsrEmailRepo.CreateMultiple")
	defer span.End()

	now := time.Now()
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	queueFn := func(b *pgx.Batch, u *entity.UsrEmail) {
		fieldsToCreate, valuesToCreate := database.GetFieldMapExcept(u, "import_id")

		createdUsrEmail := &entity.UsrEmail{}
		database.AllNullEntity(createdUsrEmail)
		createdFields, _ := createdUsrEmail.FieldMap()
		// Use DO UPDATE SET to guarantee the id will be
		// returned even the record is existed before
		stmt :=
			`
		INSERT INTO 
			%s(%s)
		VALUES 
			(%s) 
		ON CONFLICT 
			ON CONSTRAINT usr_email__pkey
		DO UPDATE SET 
			updated_at = EXCLUDED.updated_at 
		RETURNING 
			%s;
		`

		stmt = fmt.Sprintf(
			stmt,
			u.TableName(),
			strings.Join(fieldsToCreate, ","),
			database.GeneratePlaceholders(len(fieldsToCreate)),
			strings.Join(createdFields, ","),
		)

		b.Queue(stmt, valuesToCreate...)
	}

	b := &pgx.Batch{}
	for _, user := range users {
		email := strings.TrimSpace(user.LoginEmail.String)
		email = strings.ToLower(email)
		usrEmailToCreate := &entity.UsrEmail{
			UsrID:        user.ID,
			Email:        database.Text(email),
			CreatedAt:    database.Timestamptz(now),
			UpdatedAt:    database.Timestamptz(now),
			DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
			ResourcePath: database.Text(resourcePath),
		}

		queueFn(b, usrEmailToCreate)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	usrEmails := []*entity.UsrEmail{}
	for range users {
		createdUsrEmail := &entity.UsrEmail{}
		database.AllNullEntity(createdUsrEmail)
		_, createdValues := createdUsrEmail.FieldMap()
		err := batchResults.QueryRow().Scan(createdValues...)
		switch err {
		case nil:
			usrEmails = append(usrEmails, createdUsrEmail)
		case pgx.ErrNoRows:
			return nil, InternalError{
				RawError: errors.Wrap(err, "database.InsertReturning returns no row"),
			}
		default:
			return nil, InternalError{
				RawError: errors.Wrap(err, "database.InsertReturning"),
			}
		}
	}

	return usrEmails, nil
}

/*func GetByEmail(ctx context.Context, db database.QueryExecer, email pgtype.Text) (*entities.UsrEmail, error) {
	ctx, span := interceptors.StartSpan(ctx, "UsrEmailRepo.GetByEmail")
	defer span.End()

	queriedUsrEmail := &entities.UsrEmail{}
	fields, values := queriedUsrEmail.FieldMap()

	stmt :=
		`
		SELECT %s FROM %s WHERE email = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fields, ","),
		queriedUsrEmail.TableName(),
	)

	err := db.QueryRow(ctx, stmt, email).Scan(values...)

	switch err {
	case nil:
		return queriedUsrEmail, nil
	case pgx.ErrNoRows:
		return nil, errors.New("QueryRow.Scan returns no row")
	default:
		return nil, errors.Wrap(err, "QueryRow.Scan")
	}
}*/

func (r *UsrEmailRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, usrID, resourcePath, newEmail pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UsrEmailRepo.UpdateEmail")
	defer span.End()

	newEmail = database.Text(strings.ToLower(newEmail.String))

	query := fmt.Sprintf(
		`
		   UPDATE %s
		   SET email = $1
		   WHERE
		     usr_id = $2 AND
		     resource_path = $3
	       `,
		new(entity.UsrEmail).TableName(),
	)
	cmdTag, err := db.Exec(ctx, query, newEmail, usrID, resourcePath)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNoRowAffected
	}

	return nil
}
