package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainUserRepo struct{}

type User struct {
	ID                            field.String
	GroupID                       field.String
	CountryAttr                   field.String
	UserNameAttr                  field.String
	FullNameAttr                  field.String
	FirstNameAttr                 field.String
	LastNameAttr                  field.String
	FirstNamePhoneticAttr         field.String
	LastNamePhoneticAttr          field.String
	FullNamePhoneticAttr          field.String
	GivenNameAttr                 field.String
	AvatarAttr                    field.String
	PhoneNumberAttr               field.String
	EmailAttr                     field.String
	DeviceTokenAttr               field.String
	AllowNotificationAttr         field.Boolean
	LastLoginDateAttr             field.Time
	BirthdayAttr                  field.Date
	GenderAttr                    field.String
	IsTesterAttr                  field.Boolean
	FacebookIDAttr                field.String
	PhoneVerifiedAttr             field.Boolean
	EmailVerifiedAttr             field.Boolean
	OrganizationIDAttr            field.String
	ExternalUserIDAttr            field.String
	RemarksAttr                   field.String
	EncryptedUserIDByPasswordAttr field.String
	DeactivatedAtAttr             field.Time
	LoginEmailAttr                field.String
	UserRoleAttr                  field.String

	// These attributes belong to postgres database context
	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func NewUser(user entity.User) (*User, error) {
	now := field.NewTime(time.Now())

	// Get encrypted value of user id
	encryptedUserIDByPassword, err := entity.EncryptedUserIDByPasswordFromUser(user)
	if err != nil {
		return nil, err
	}

	repoUser := &User{
		ID:                            user.UserID(),
		GroupID:                       user.Group(),
		UserNameAttr:                  user.UserName(),
		CountryAttr:                   user.Country(),
		FullNameAttr:                  user.FullName(),
		FirstNameAttr:                 user.FirstName(),
		LastNameAttr:                  user.LastName(),
		FirstNamePhoneticAttr:         user.FirstNamePhonetic(),
		LastNamePhoneticAttr:          user.LastNamePhonetic(),
		FullNamePhoneticAttr:          user.FullNamePhonetic(),
		GivenNameAttr:                 user.GivenName(),
		AvatarAttr:                    user.Avatar(),
		PhoneNumberAttr:               user.PhoneNumber(),
		EmailAttr:                     user.Email(),
		DeviceTokenAttr:               user.DeviceToken(),
		AllowNotificationAttr:         user.AllowNotification(),
		LastLoginDateAttr:             user.LastLoginDate(),
		BirthdayAttr:                  user.Birthday(),
		GenderAttr:                    user.Gender(),
		IsTesterAttr:                  user.IsTester(),
		FacebookIDAttr:                user.FacebookID(),
		PhoneVerifiedAttr:             user.PhoneVerified(),
		EmailVerifiedAttr:             user.EmailVerified(),
		OrganizationIDAttr:            user.OrganizationID(),
		ExternalUserIDAttr:            user.ExternalUserID(),
		CreatedAt:                     now,
		UpdatedAt:                     now,
		DeletedAt:                     field.NewNullTime(),
		DeactivatedAtAttr:             user.DeactivatedAt(),
		RemarksAttr:                   user.Remarks(),
		LoginEmailAttr:                user.LoginEmail(),
		EncryptedUserIDByPasswordAttr: encryptedUserIDByPassword,
		UserRoleAttr:                  user.UserRole(),
	}
	field.SetUndefinedFieldsToNull(repoUser)
	return repoUser, nil
}

func (u *User) UserID() field.String {
	return u.ID
}
func (u *User) Avatar() field.String {
	return u.AvatarAttr
}
func (u *User) Group() field.String {
	return u.GroupID
}
func (u *User) UserName() field.String {
	return u.UserNameAttr
}
func (u *User) FullName() field.String {
	return u.FullNameAttr
}
func (u *User) FirstName() field.String {
	return u.FirstNameAttr
}
func (u *User) LastName() field.String {
	return u.LastNameAttr
}
func (u *User) GivenName() field.String {
	return u.GivenNameAttr
}
func (u *User) FullNamePhonetic() field.String {
	return u.FullNamePhoneticAttr
}
func (u *User) FirstNamePhonetic() field.String {
	return u.FirstNamePhoneticAttr
}
func (u *User) LastNamePhonetic() field.String {
	return u.LastNamePhoneticAttr
}
func (u *User) Country() field.String {
	return u.CountryAttr
}
func (u *User) PhoneNumber() field.String {
	return u.PhoneNumberAttr
}
func (u *User) Email() field.String {
	return u.EmailAttr
}
func (u *User) DeviceToken() field.String {
	return u.DeviceTokenAttr
}
func (u *User) AllowNotification() field.Boolean {
	return u.AllowNotificationAttr
}
func (u *User) LastLoginDate() field.Time {
	return u.LastLoginDateAttr
}
func (u *User) Birthday() field.Date {
	return u.BirthdayAttr
}
func (u *User) Gender() field.String {
	return u.GenderAttr
}
func (u *User) IsTester() field.Boolean {
	return u.IsTesterAttr
}
func (u *User) FacebookID() field.String {
	return u.FacebookIDAttr
}
func (u *User) ExternalUserID() field.String {
	return u.ExternalUserIDAttr
}
func (u *User) PhoneVerified() field.Boolean {
	return u.PhoneVerifiedAttr
}
func (u *User) EmailVerified() field.Boolean {
	return u.EmailVerifiedAttr
}
func (u *User) Password() field.String {
	return field.NewNullString()
}
func (u *User) OrganizationID() field.String {
	return u.OrganizationIDAttr
}
func (u *User) Remarks() field.String {
	return u.RemarksAttr
}
func (u *User) EncryptedUserIDByPassword() field.String {
	return u.EncryptedUserIDByPasswordAttr
}
func (u *User) DeactivatedAt() field.Time {
	return u.DeactivatedAtAttr
}
func (u *User) LoginEmail() field.String {
	return u.LoginEmailAttr
}
func (u *User) UserRole() field.String {
	return u.UserRoleAttr
}

// FieldMap returns field in users table
func (u *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"country",
			"username",
			"name",
			"first_name",
			"last_name",
			"first_name_phonetic",
			"last_name_phonetic",
			"full_name_phonetic",
			"given_name",
			"avatar",
			"phone_number",
			"email",
			"device_token",
			"allow_notification",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"last_login_date",
			"birthday",
			"gender",
			"is_tester",
			"facebook_id",
			"phone_verified",
			"email_verified",
			"user_external_id",
			"remarks",
			"encrypted_user_id_by_password",
			"login_email",
			"user_role",
		}, []interface{}{
			&u.ID,
			&u.GroupID,
			&u.CountryAttr,
			&u.UserNameAttr,
			&u.FullNameAttr,
			&u.FirstNameAttr,
			&u.LastNameAttr,
			&u.FirstNamePhoneticAttr,
			&u.LastNamePhoneticAttr,
			&u.FullNamePhoneticAttr,
			&u.GivenNameAttr,
			&u.AvatarAttr,
			&u.PhoneNumberAttr,
			&u.EmailAttr,
			&u.DeviceTokenAttr,
			&u.AllowNotificationAttr,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
			&u.OrganizationIDAttr,
			&u.LastLoginDateAttr,
			&u.BirthdayAttr,
			&u.GenderAttr,
			&u.IsTesterAttr,
			&u.FacebookIDAttr,
			&u.PhoneVerifiedAttr,
			&u.EmailVerifiedAttr,
			&u.ExternalUserIDAttr,
			&u.RemarksAttr,
			&u.EncryptedUserIDByPasswordAttr,
			&u.LoginEmailAttr,
			&u.UserRoleAttr,
		}
}

func (u *User) TableName() string {
	return "users"
}

func (repo *DomainUserRepo) create(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.Create")
	defer span.End()

	databaseUserToCreate, err := NewUser(userToCreate)
	if err != nil {
		return err
	}

	cmdTag, err := database.Insert(ctx, databaseUserToCreate, db.Exec)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return ErrNoRowAffected
	}

	return nil
}

func (repo *DomainUserRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.UpsertMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, user *User) {
		fields, values := user.FieldMap()

		insertPlaceHolders := database.GeneratePlaceholders(len(fields))
		updatePlaceHolders := generateUpdateUserPlaceholders(fields, isEnableUsername)
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT users_pk DO UPDATE SET %s",
			user.TableName(),
			strings.Join(fields, ","),
			insertPlaceHolders,
			updatePlaceHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, userToCreate := range usersToCreate {
		repoDomainUser, err := NewUser(userToCreate)
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "failed to init repo enity")}
		}
		queueFn(batch, repoDomainUser)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(usersToCreate); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}

		if cmdTag.RowsAffected() != 1 {
			return InternalError{RawError: fmt.Errorf("user was not inserted")}
		}
	}

	return nil
}

func (repo *DomainUserRepo) GetByUserNames(ctx context.Context, db database.QueryExecer, usernames []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.GetByUserNames")
	defer span.End()

	user := new(User)
	fields, _ := user.FieldMap()

	query := fmt.Sprintf(
		`
			SELECT %s FROM %s 
			WHERE username = ANY($1)
		`,
		strings.Join(fields, ","), user.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(usernames))
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()
	if rows.Err() != nil {
		return nil, InternalError{errors.Wrap(rows.Err(), "rows.Err")}
	}

	users := make([]entity.User, 0, len(usernames))
	for rows.Next() {
		item := new(User)
		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		users = append(users, item)
	}

	return users, nil
}

func (repo *DomainUserRepo) GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetByEmails")
	defer span.End()

	user := &User{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (email = ANY($1))",
		strings.Join(fields, ","), user.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(emails),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	users := make([]entity.User, 0, len(emails))
	for rows.Next() {
		item := &User{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		users = append(users, item)
	}

	return users, nil
}

func (repo *DomainUserRepo) GetByEmailsInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.GetByEmailsInsensitiveCase")
	defer span.End()

	user := &User{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (email = ANY($1))",
		strings.Join(fields, ","), user.TableName())

	// lowerCaseEmails := make([]string, 0, len(emails))
	// for _, email := range emails {
	// 	lowerCaseEmails = append(lowerCaseEmails, strings.ToLower(email))
	// }

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(emails),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}

	defer rows.Close()

	users := make([]entity.User, 0, len(emails))
	for rows.Next() {
		user := &User{}

		_, fieldValues := user.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "rows.Scan"),
			}
		}

		users = append(users, user)
	}

	return users, nil
}

func (repo *DomainUserRepo) UpdateEmail(ctx context.Context, db database.QueryExecer, usersToUpdate entity.User) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.UpdateEmail")
	defer span.End()

	fields := []string{
		"email",
	}

	user, err := NewUser(usersToUpdate)
	if err != nil {
		return err
	}

	cmdTag, err := database.UpdateFields(ctx, user, db.Exec, "user_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNoRowAffected
	}
	return nil
}

func (repo *DomainUserRepo) GetByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.GetByIDs")
	defer span.End()

	user := &User{}
	fields, _ := user.FieldMap()

	query := fmt.Sprintf(
		`
			SELECT %s
			FROM %s
			WHERE
				user_id = ANY($1) AND
				deleted_at IS NULL
		`,
		strings.Join(fields, ","), user.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(userIDs),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}

	defer rows.Close()

	users := make([]entity.User, 0, len(userIDs))
	for rows.Next() {
		item := &User{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: errors.Wrap(err, "rows.Scan"),
			}
		}

		users = append(users, item)
	}

	return users, nil
}

func (repo *DomainUserRepo) GetByExternalUserIDs(ctx context.Context, db database.QueryExecer, partnerInternalIDs []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.GetByExternalUserIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE user_external_id = ANY($1) AND deleted_at IS NULL`

	user, err := NewUser(entity.EmptyUser{})
	if err != nil {
		return nil, InternalError{
			RawError: errors.WithStack(err),
		}
	}

	fieldNames, _ := user.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		user.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(partnerInternalIDs),
	)
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}

	defer rows.Close()

	var result entity.Users
	for rows.Next() {
		item, err := NewUser(entity.EmptyUser{})
		if err != nil {
			return nil, InternalError{
				RawError: errors.WithStack(err),
			}
		}

		_, fieldValues := item.FieldMap()

		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: fmt.Errorf("rows.Scan: %w", err),
			}
		}

		result = append(result, item)
	}
	return result, nil
}

func (repo *DomainUserRepo) UpdateActivation(ctx context.Context, db database.QueryExecer, users entity.Users) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.UpdateActivation")
	defer span.End()

	queueFn := func(batch *pgx.Batch, user *User) {
		stmt := `
			UPDATE users
			SET
				deactivated_at = $1
			WHERE
				user_id = $2
		`
		batch.Queue(stmt, &user.DeactivatedAtAttr, &user.ID)
	}

	batch := &pgx.Batch{}
	for _, user := range users {
		domainUser, err := NewUser(user)
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "failed to init repo enity")}
		}
		queueFn(batch, domainUser)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for range users {
		if _, err := batchResults.Exec(); err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}
	}

	return nil
}

func (repo *DomainUserRepo) GetUserRoles(ctx context.Context, db database.QueryExecer, userID string) (entity.DomainRoles, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserRepo.GetUserRoles")
	defer span.End()

	fields, _ := (&entity.Role{}).FieldMap()

	stmt := fmt.Sprintf(
		`SELECT r.%s FROM user_group_member ugm
			INNER JOIN granted_role gt ON ugm.user_group_id = gt.user_group_id
			INNER JOIN role r ON gt.role_id = r.role_id
		WHERE ugm.user_id = $1
			AND gt.deleted_at IS NULL
			AND ugm.deleted_at IS NULL
			AND r.deleted_at IS NULL`, strings.Join(fields, ", r."))

	rows, err := db.Query(
		ctx,
		stmt,
		database.Text(userID),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := entity.DomainRoles{}
	for rows.Next() {
		item := NewRole(entity.NullDomainRole{})

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		result = append(result, item)
	}
	return result, nil
}

func generateUpdateUserPlaceholders(fields []string, isEnableUsername bool) string {
	var builder strings.Builder
	sep := ", "
	ignoreFields := []string{"created_at", "deleted_at", "device_token", "avatar", "allow_notification", "last_login_date", "is_tester", "facebook_id", "phone_verified", "email_verified"}
	if isEnableUsername {
		ignoreFields = append(ignoreFields, "login_email")
	}
	filteredFields := make([]string, 0, len(fields))

	for _, field := range fields {
		if utils.IndexOf(ignoreFields, field) == -1 {
			filteredFields = append(filteredFields, field)
		}
	}
	totalField := len(filteredFields)

	for i, field := range filteredFields {
		if i == totalField-1 {
			sep = ""
		}

		builder.WriteString(field + " = EXCLUDED." + field + sep)
	}

	return builder.String()
}
