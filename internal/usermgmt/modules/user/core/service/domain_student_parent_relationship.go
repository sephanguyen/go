package service

import (
	"context"

	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type StudentParentRelationShipRepo interface {
	AssignParentsToStudent(ctx context.Context, db libdatabase.QueryExecer, orgID valueobj.HasOrganizationID, relationship field.String, studentIDToBeAssigned valueobj.HasUserID, parentIDsToAssign ...valueobj.HasUserID) error
	AssignParentForStudents(ctx context.Context, db libdatabase.QueryExecer, orgID valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error
}

type StudentParentRelationshipManager func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, studentIDToBeAssigned valueobj.HasUserID, parentIDsToAssign ...valueobj.HasUserID) error

type AssignParentToStudentsManager func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error

func NewStudentParentRelationshipManager(studentParentRelationShipRepo StudentParentRelationShipRepo) StudentParentRelationshipManager {
	return func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, studentIDToBeAssigned valueobj.HasUserID, parentIDsToAssign ...valueobj.HasUserID) error {
		err := errorx.ReturnFirstErr(
			studentParentRelationShipRepo.AssignParentsToStudent(ctx, db, org, relationship, studentIDToBeAssigned, parentIDsToAssign...),
		)
		return err
	}
}

func NewAssignParentToStudentsManager(studentParentRelationShipRepo StudentParentRelationShipRepo) AssignParentToStudentsManager {
	return func(ctx context.Context, db libdatabase.QueryExecer, org valueobj.HasOrganizationID, relationship field.String, parentIDtoBeAssigned valueobj.HasUserID, studentIDsToBeAssigned ...valueobj.HasStudentID) error {
		err := errorx.ReturnFirstErr(
			studentParentRelationShipRepo.AssignParentForStudents(ctx, db, org, relationship, parentIDtoBeAssigned, studentIDsToBeAssigned...),
		)
		return err
	}
}
