package common

// notification, user, student, parent, location, course, student_course, class

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

func (s *suite) insertPermissionRoleForNotificationEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()
	ownerPermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'communication.notification.read', now(), now(), autofillresourcepath()),
			($2, 'communication.notification.write', now(), now(), autofillresourcepath()),
			($3, 'communication.notification.owner', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID, ownerPermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),
		($1, $9, now(), now(), autofillresourcepath()),
		($1, $10, now(), now(), autofillresourcepath()),

		($2, $7, now(), now(), autofillresourcepath()),
		($2, $8, now(), now(), autofillresourcepath()),
	
		($3, $6, now(), now(), autofillresourcepath()),
		($3, $7, now(), now(), autofillresourcepath()),
		($3, $8, now(), now(), autofillresourcepath()),
		($3, $9, now(), now(), autofillresourcepath()),
		($3, $10, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID, ownerPermissionID,
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForLocationEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'master.location.read', now(), now(), autofillresourcepath());
	`, readPermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $2, now(), now(), autofillresourcepath()),
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),
		($1, $9, now(), now(), autofillresourcepath()),
		($1, $10, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID,
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleTeacherLead],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForStudentEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readStudentPermissionID := idutil.ULIDNow()
	writeStudentPermissionID := idutil.ULIDNow()
	readStudentPaymentDetailPermissionID := idutil.ULIDNow()
	writeStudentPaymentDetailPermissionID := idutil.ULIDNow()
	readStudentCoursePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'user.student.read', now(), now(), autofillresourcepath()),
			($2, 'user.student.write', now(), now(), autofillresourcepath()),
			($3, 'user.studentpaymentdetail.read', now(), now(), autofillresourcepath()),
			($4, 'user.studentpaymentdetail.write', now(), now(), autofillresourcepath()),
			($5, 'user.student_course.write', now(), now(), autofillresourcepath())
		ON CONFLICT DO NOTHING;
	`, readStudentPermissionID, writeStudentPermissionID, readStudentPaymentDetailPermissionID, writeStudentPaymentDetailPermissionID, readStudentCoursePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
		INSERT INTO permission_role 
			(permission_id, role_id, created_at, updated_at, resource_path)
		VALUES 
			($1, $6, now(), now(), autofillresourcepath()),
			($1, $7, now(), now(), autofillresourcepath()),
			($1, $8, now(), now(), autofillresourcepath()),
			($1, $9, now(), now(), autofillresourcepath()),
			($1, $10, now(), now(), autofillresourcepath()),
			($1, $11, now(), now(), autofillresourcepath()),

			($2, $7, now(), now(), autofillresourcepath()),
			($2, $8, now(), now(), autofillresourcepath()),
			($2, $9, now(), now(), autofillresourcepath()),
			($2, $10, now(), now(), autofillresourcepath()),
			($2, $11, now(), now(), autofillresourcepath()),

			($3, $7, now(), now(), autofillresourcepath()),
			($3, $8, now(), now(), autofillresourcepath()),

			($4, $7, now(), now(), autofillresourcepath()),
			($4, $8, now(), now(), autofillresourcepath()),

			($5, $7, now(), now(), autofillresourcepath()),
			($5, $8, now(), now(), autofillresourcepath()),
			($5, $10, now(), now(), autofillresourcepath()),
			($5, $11, now(), now(), autofillresourcepath())
		ON CONFLICT DO NOTHING;
	`, readStudentPermissionID, writeStudentPermissionID, readStudentPaymentDetailPermissionID, writeStudentPaymentDetailPermissionID, readStudentCoursePermissionID,
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForParentEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'user.parent.read', now(), now(), autofillresourcepath()),
			($2, 'user.parent.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),

		($2, $4, now(), now(), autofillresourcepath()),
		($2, $5, now(), now(), autofillresourcepath()),
		($2, $6, now(), now(), autofillresourcepath()),
		($2, $7, now(), now(), autofillresourcepath()),
		($2, $8, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID,
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForStaffEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'user.staff.read', now(), now(), autofillresourcepath()),
			($2, 'user.staff.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),

		($2, $3, now(), now(), autofillresourcepath()),
		($2, $4, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID,
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForUserGroupEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()
	writeUGMemberPermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'user.usergroup.read', now(), now(), autofillresourcepath()),
			($3, 'user.usergroup.write', now(), now(), autofillresourcepath()),
			($2, 'user.usergroupmember.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID, writeUGMemberPermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),

		($2, $4, now(), now(), autofillresourcepath()),
		($2, $5, now(), now(), autofillresourcepath()),

		($3, $4, now(), now(), autofillresourcepath()),
		($3, $5, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID, writeUGMemberPermissionID,
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForUserEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'user.user.read', now(), now(), autofillresourcepath()),
			($2, 'user.user.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),
		($1, $9, now(), now(), autofillresourcepath()),
		($1, $10, now(), now(), autofillresourcepath()),
		($1, $11, now(), now(), autofillresourcepath()),

		($2, $4, now(), now(), autofillresourcepath()),
		($2, $5, now(), now(), autofillresourcepath()),
		($2, $6, now(), now(), autofillresourcepath()),
		($2, $7, now(), now(), autofillresourcepath()),
		($2, $8, now(), now(), autofillresourcepath()),
		($2, $10, now(), now(), autofillresourcepath()),
		($2, $11, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID,
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleTeacherLead],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForCourseEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'master.course.read', now(), now(), autofillresourcepath()),
			($2, 'master.course.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),
		($1, $9, now(), now(), autofillresourcepath()),
		($1, $10, now(), now(), autofillresourcepath()),
		($1, $11, now(), now(), autofillresourcepath()),

		($2, $6, now(), now(), autofillresourcepath()),
		($2, $7, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID,
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleTeacherLead],
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) insertPermissionRoleForLessonEnt(ctx context.Context, mapRoleAndRoleID map[string]string, resourcePath string) error {
	ctx2 := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
		},
	})

	readPermissionID := idutil.ULIDNow()
	writePermissionID := idutil.ULIDNow()

	_, err := s.BobDB.Exec(ctx2, `
		INSERT INTO permission 
			(permission_id, permission_name, created_at, updated_at, resource_path)
		VALUES 
			($1, 'lesson.lesson.read', now(), now(), autofillresourcepath()),
			($2, 'lesson.lesson.write', now(), now(), autofillresourcepath());
	`, readPermissionID, writePermissionID)
	if err != nil {
		return err
	}

	_, err = s.BobDB.Exec(ctx2, `
	INSERT INTO permission_role 
		(permission_id, role_id, created_at, updated_at, resource_path)
	VALUES 
		($1, $3, now(), now(), autofillresourcepath()),
		($1, $4, now(), now(), autofillresourcepath()),
		($1, $5, now(), now(), autofillresourcepath()),
		($1, $6, now(), now(), autofillresourcepath()),
		($1, $7, now(), now(), autofillresourcepath()),
		($1, $8, now(), now(), autofillresourcepath()),
		($1, $9, now(), now(), autofillresourcepath()),
		($1, $10, now(), now(), autofillresourcepath()),
		($1, $11, now(), now(), autofillresourcepath()),

		($2, $3, now(), now(), autofillresourcepath()),
		($2, $5, now(), now(), autofillresourcepath()),
		($2, $6, now(), now(), autofillresourcepath()),
		($2, $7, now(), now(), autofillresourcepath()),
		($2, $8, now(), now(), autofillresourcepath()),
		($2, $9, now(), now(), autofillresourcepath())
	ON CONFLICT DO NOTHING;
	`, readPermissionID, writePermissionID,
		mapRoleAndRoleID[constant.RoleStudent],
		mapRoleAndRoleID[constant.RoleParent],
		mapRoleAndRoleID[constant.RoleTeacher],
		mapRoleAndRoleID[constant.RoleSchoolAdmin],
		mapRoleAndRoleID[constant.RoleHQStaff],
		mapRoleAndRoleID[constant.RoleCentreManager],
		mapRoleAndRoleID[constant.RoleCentreStaff],
		mapRoleAndRoleID[constant.RoleCentreLead],
		mapRoleAndRoleID[constant.RoleTeacherLead],
	)
	if err != nil {
		return err
	}

	return nil
}
