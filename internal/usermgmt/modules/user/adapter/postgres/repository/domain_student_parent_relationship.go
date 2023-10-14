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

type DomainStudentParentRelationshipRepo struct {
}

type DomainStudentParentRelationship struct {
	StudentIDAttr    field.String
	ParentIDAttr     field.String
	RelationshipAttr field.String
	ResourcePathAttr field.String
	UpdatedAtAttr    field.Time
	CreatedAtAttr    field.Time
	DeletedAtAttr    field.Time
}

func (relationship *DomainStudentParentRelationship) StudentID() field.String {
	return relationship.StudentIDAttr
}

func (relationship *DomainStudentParentRelationship) ParentID() field.String {
	return relationship.ParentIDAttr
}

func (relationship *DomainStudentParentRelationship) Relationship() field.String {
	return relationship.RelationshipAttr
}

func (relationship *DomainStudentParentRelationship) OrganizationID() field.String {
	return relationship.ResourcePathAttr
}

func (relationship *DomainStudentParentRelationship) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_id", "parent_id", "relationship", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&relationship.StudentIDAttr, &relationship.ParentIDAttr, &relationship.RelationshipAttr, &relationship.CreatedAtAttr, &relationship.UpdatedAtAttr, &relationship.DeletedAtAttr}
	return
}

func (relationship *DomainStudentParentRelationship) TableName() string {
	return "student_parents"
}

// UserIDDelegated NewStudentParentRelationship receive studentID valueobj.HasUserID and parentID valueobj.HasUserID
// this one to convert valueobj.HasStudentID to valueobj.HasUserID
// discussion thread: https://github.com/manabie-com/backend/pull/14271#discussion_r1165276665
// TODO: will remove when we can find a better way to convert valueobj.HasStudentID to valueobj.HasUserID
type UserIDDelegated struct {
	StudentIDAttr field.String
}

func (u UserIDDelegated) UserID() field.String {
	return u.StudentIDAttr
}

func NewStudentParentRelationship(orgID valueobj.HasOrganizationID, relationship field.String, studentID valueobj.HasUserID, parentID valueobj.HasUserID) *DomainStudentParentRelationship {
	now := field.NewTime(time.Now())

	return &DomainStudentParentRelationship{
		StudentIDAttr:    studentID.UserID(),
		ParentIDAttr:     parentID.UserID(),
		RelationshipAttr: relationship,
		ResourcePathAttr: orgID.OrganizationID(),
		UpdatedAtAttr:    now,
		CreatedAtAttr:    now,
		DeletedAtAttr:    field.NewNullTime(),
	}
}

func (repo *DomainStudentParentRelationshipRepo) AssignParentsToStudent(ctx context.Context, db database.QueryExecer, orgID valueobj.HasOrganizationID, relationship field.String, studentIDToBeAssigned valueobj.HasUserID, parentIDsToAssign ...valueobj.HasUserID) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.AssignParentsToStudent")
	defer span.End()

	b := &pgx.Batch{}
	for _, parentIDToAssign := range parentIDsToAssign {
		studentParentRelationshipToUpsert := NewStudentParentRelationship(orgID, relationship, studentIDToBeAssigned, parentIDToAssign)
		fieldNames, fieldValues := studentParentRelationshipToUpsert.FieldMap()
		insertPlaceHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt :=
			`
			INSERT INTO 
			     student_parents (%s) 
			VALUES 
			     (%s)
			ON CONFLICT ON CONSTRAINT student_parents_pk
				DO UPDATE SET 
					relationship = $3, updated_at = $5, deleted_at = NULL
			`

		stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), insertPlaceHolders)

		b.Queue(stmt, fieldValues...)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (repo *DomainStudentParentRelationshipRepo) SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.SoftDeleteByStudentIDs")
	defer span.End()

	studentParent := DomainStudentParentRelationship{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE student_id = ANY($1)`, studentParent.TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(studentIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainStudentParentRelationshipRepo) SoftDeleteByParentIDs(ctx context.Context, db database.QueryExecer, parentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.SoftDeleteByParentIDs")
	defer span.End()

	studentParent := DomainStudentParentRelationship{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE parent_id = ANY($1)`, studentParent.TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(parentIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainStudentParentRelationshipRepo) AssignParentForStudents(ctx context.Context, db database.QueryExecer, orgID valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.AssignParentForStudents")
	defer span.End()

	b := &pgx.Batch{}

	for _, studentIDToBeAssigned := range studentIDsToBeAssigned {
		studentParentRelationshipToUpsert := NewStudentParentRelationship(orgID, relationship, UserIDDelegated{
			StudentIDAttr: studentIDToBeAssigned.StudentID(),
		}, parentIDtoBeAssigned)
		fieldNames, fieldValues := studentParentRelationshipToUpsert.FieldMap()
		insertPlaceHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt :=
			`
			INSERT INTO 
			     student_parents (%s) 
			VALUES 
			     (%s)
			ON CONFLICT ON CONSTRAINT student_parents_pk
				DO UPDATE SET 
					relationship = $3, updated_at = $5, deleted_at = NULL
			`

		stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), insertPlaceHolders)

		b.Queue(stmt, fieldValues...)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (repo *DomainStudentParentRelationshipRepo) GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (entity.DomainStudentParentRelationships, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.GetByStudentIDs")
	defer span.End()

	studentParent := &DomainStudentParentRelationship{}
	fields, _ := studentParent.FieldMap()
	stmt := `SELECT %s FROM %s WHERE student_id = ANY($1) AND deleted_at IS NULL`
	query := fmt.Sprintf(stmt, strings.Join(fields, ","), studentParent.TableName())

	rows, err := db.Query(ctx, query, database.TextArray(studentIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentParents := make(entity.DomainStudentParentRelationships, 0)
	for rows.Next() {
		studentParent := &DomainStudentParentRelationship{}
		err := rows.Scan(database.GetScanFields(studentParent, database.GetFieldNames(studentParent))...)
		if err != nil {
			return nil, err
		}
		studentParents = append(studentParents, studentParent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return studentParents, nil
}

func (repo *DomainStudentParentRelationshipRepo) GetByParentIDs(ctx context.Context, db database.QueryExecer, parentIDs []string) (entity.DomainStudentParentRelationships, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentParentRelationshipRepo.GetByParentIDs")
	defer span.End()

	studentParent := &DomainStudentParentRelationship{}
	fields, _ := studentParent.FieldMap()
	stmt := `SELECT %s FROM %s WHERE parent_id = ANY($1) AND deleted_at IS NULL`
	query := fmt.Sprintf(stmt, strings.Join(fields, ","), studentParent.TableName())

	rows, err := db.Query(ctx, query, database.TextArray(parentIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentParents := make(entity.DomainStudentParentRelationships, 0)
	for rows.Next() {
		studentParent := &DomainStudentParentRelationship{}
		err := rows.Scan(database.GetScanFields(studentParent, database.GetFieldNames(studentParent))...)
		if err != nil {
			return nil, err
		}
		studentParents = append(studentParents, studentParent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return studentParents, nil
}
