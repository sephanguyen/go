package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainUsrEmailRepo struct{}

type UsrEmail struct {
	UsrID        field.String
	Email        field.String
	CreatedAt    field.Time
	UpdatedAt    field.Time
	DeletedAt    field.Time
	ResourcePath field.String
	ImportID     field.Int64
}

func (e UsrEmail) UserID() field.String {
	return e.UsrID
}

func (e *UsrEmail) TableName() string {
	return "usr_email"
}

func (e *UsrEmail) FieldMap() ([]string, []interface{}) {
	return []string{
			"usr_id",
			"email",
			"create_at",
			"updated_at",
			"delete_at",
			"resource_path",
			"import_id",
		}, []interface{}{
			&e.UsrID,
			&e.Email,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
			&e.ImportID,
		}
}

func (r *DomainUsrEmailRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, user entity.User) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUsrEmailRepo.UpdateEmail")
	defer span.End()

	email := strings.TrimSpace(user.LoginEmail().String())
	email = strings.ToLower(email)
	query := fmt.Sprintf(
		`
		   	UPDATE 
		       %s
		   	SET 
		    	email = $1
		   	WHERE
				usr_id = $2
	       `,
		new(UsrEmail).TableName(),
	)
	_, err := db.Exec(ctx, query, email, user.UserID())
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "db.Exec")}
	}

	return nil
}

func (r *DomainUsrEmailRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, users entity.Users) (valueobj.HasUserIDs, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUsrEmailRepo.CreateMultiple")
	defer span.End()

	now := time.Now()

	queueFn := func(b *pgx.Batch, u *UsrEmail) {
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
		email := strings.TrimSpace(user.LoginEmail().String())
		email = strings.ToLower(email)
		usrEmailToCreate := &UsrEmail{
			UsrID:        user.UserID(),
			Email:        field.NewString(email),
			CreatedAt:    field.NewTime(now),
			UpdatedAt:    field.NewTime(now),
			DeletedAt:    field.NewNullTime(),
			ResourcePath: user.OrganizationID(),
		}

		queueFn(b, usrEmailToCreate)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	userIDs := []valueobj.HasUserID{}
	for range users {
		createdUsrEmail := UsrEmail{}
		_, createdValues := createdUsrEmail.FieldMap()
		err := batchResults.QueryRow().Scan(createdValues...)
		switch err {
		case nil:
			userIDs = append(userIDs, createdUsrEmail)
		case pgx.ErrNoRows:
			return nil, InternalError{RawError: errors.Wrap(err, "database.InsertReturning returns no row")}
		default:
			return nil, InternalError{RawError: errors.Wrap(err, "database.InsertReturning")}
		}
	}

	return userIDs, nil
}
