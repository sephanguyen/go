package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	listPermission map[string]string
	listRole       map[string]string
)

func main() {
	// Set to false if don't want to generate table related user: users, school_admins, users_groups, staff, user_group_member, granted_permission
	generateUserTables := true
	resourcePath := "100011"
	tenantID := "e2e-adobo-vgmt0"
	tenantName := "E2E adobo"
	domainName := "e2e-adobo"

	locationID := ULIDNow()
	locationTypeID := ULIDNow()

	orgStmt := generateOrgScript(resourcePath, tenantID, tenantName, domainName)
	locationStmt := generateLocationScript(resourcePath, tenantName, locationID, locationTypeID)
	schoolStmt := generateSchoolScript(resourcePath, tenantName, tenantID)
	permissionStmt := generatePermissionScript(resourcePath)
	roleStmt := generateRoleScript(resourcePath)
	permissionRoleStmt := generatePermissionRoleScript(resourcePath)
	userGroupStmt, groupIDs := generateUserGroupScript(resourcePath)
	grantedRoleStmt, grantedRoleIDs := generateGrantedRoleScript(resourcePath, groupIDs)
	grantedRoleACStmt := generateGrantedRoleAccessPathScript(resourcePath, grantedRoleIDs, locationID)
	finalScript := orgStmt + "\n\n" + locationStmt + "\n\n" + schoolStmt + "\n\n" + permissionStmt + "\n\n" + roleStmt + "\n\n" + permissionRoleStmt + "\n\n" + userGroupStmt + "\n\n" + grantedRoleStmt + "\n\n" + grantedRoleACStmt
	// only generate for user belongs to group school admin
	if generateUserTables {
		// get from identity platform
		email := "loc.nguyen+e2eadoboadmin@manabie.com"
		userID := "Gaw8oce0elM8IHu5Q8V8bxVRiI73"
		userStmt := generateUserScript(resourcePath, email, userID, groupIDs["School Admin"], locationID)
		finalScript = finalScript + "\n\n" + userStmt
	}

	f, err := os.Create("./script-new-org.sql")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.WriteString(finalScript)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
}

func generateUserScript(resourcePath, email, userID, userGroupID, locationID string) string {
	iresourcePath, _ := strconv.Atoi(resourcePath)
	stmt := fmt.Sprintf(`-- insert Users
INSERT INTO public.users
	(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id)
	VALUES ('%s', 'COUNTRY_JP', 'School Admin', '', NULL, '%s', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '%s', NULL, NULL, NULL, '', '', NULL, NULL, NULL, NULL, false, NULL) ON CONFLICT DO NOTHING;
-- insert school_admins
INSERT INTO public.school_admins
	(school_admin_id, school_id, updated_at, created_at, deleted_at, resource_path)
	VALUES('%s', %d, now(), now(), NULL, '%s')
	ON CONFLICT DO NOTHING;
-- insert users_groups
INSERT INTO public.users_groups
	(user_id, group_id, is_origin, status, updated_at, created_at, resource_path)
	VALUES('%s', 'USER_GROUP_SCHOOL_ADMIN', true, 'USER_GROUP_STATUS_ACTIVE', now(), now(), '%s')
	ON CONFLICT DO NOTHING;
-- insert staff
INSERT INTO public.staff
	(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
	VALUES('%s', now(), now(), NULL, '%s', false, 'AVAILABLE', NULL, NULL)
	ON CONFLICT DO NOTHING;
-- insert user_group_member
INSERT INTO public.user_group_member
	(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
	VALUES('%s', '%s', now(), now(), NULL, '%s')
	ON CONFLICT DO NOTHING;
-- insert granted_permission
INSERT INTO granted_permission
	(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
	SELECT * FROM retrieve_src_granted_permission('%s')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;
-- insert user_access_paths
INSERT INTO public.user_access_paths
	(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
	VALUES('%s', '%s', NULL, now(), now(), NULL, '%s')
	ON CONFLICT DO NOTHING;`, userID, email, resourcePath, userID, iresourcePath, resourcePath, userID, resourcePath, userID, resourcePath, userID, userGroupID, resourcePath, userGroupID, userID, locationID, resourcePath)
	return stmt
}

func generateGrantedRoleAccessPathScript(resourcePath string, grantedRoleIDs []string, locationID string) string {
	stmt := ` -- insert granted_role_access_path
INSERT INTO public.granted_role_access_path
	(granted_role_id, location_id, created_at, updated_at, resource_path)
  VALUES`

	for _, grantedRoleID := range grantedRoleIDs {
		stmt = fmt.Sprintf("%s\n ('%s', '%s', now(), now(), '%s'),", stmt, grantedRoleID, locationID, resourcePath)
	}

	stmt = stmt[:len(stmt)-1] + ";"
	return stmt
}

func generateGrantedRoleScript(resourcePath string, userGroup map[string]string) (string, []string) {
	grantedRoleIDs := []string{}
	listRoleIDs := ListRoleIDs()
	stmt := ` -- insert granted_role
INSERT INTO public.granted_role
	(granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
  VALUES`

	for groupname, groupID := range userGroup {
		roleID, ok := listRoleIDs[groupname]
		if ok {
			id := ULIDNow()
			grantedRoleIDs = append(grantedRoleIDs, id)
			stmt = fmt.Sprintf("%s\n ('%s', '%s', '%s', now(), now(), '%s'),", stmt, id, groupID, roleID, resourcePath)
		}
	}

	stmt = stmt[:len(stmt)-1] + ";"
	if len(grantedRoleIDs) == 0 {
		return "", grantedRoleIDs
	}
	return stmt, grantedRoleIDs
}

func generateUserGroupScript(resourcePath string) (string, map[string]string) {
	groupIDs := map[string]string{}
	stmt := ` -- insert user group
INSERT INTO public.user_group
	(user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES`

	for _, groupname := range listgroupName {
		id := ULIDNow()
		groupIDs[groupname] = id
		stmt = fmt.Sprintf("%s\n ('%s', '%s', true, now(), now(), '%s'),", stmt, id, groupname, resourcePath)
	}

	stmt = stmt[:len(stmt)-1] + ";"
	return stmt, groupIDs
}

func generateOrgScript(resourcePath, tenantID, name, domain string) string {
	stmt := fmt.Sprintf(`-- insert Organization on mastermgmt DB
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
	VALUES ('%s', '%s', '%s', '%s', '%s', '', 'COUNTRY_JP', now(), now(), null)`, resourcePath, tenantID, name, resourcePath, domain)
	return stmt
}

func generateLocationScript(resourcePath, name, locationID, locationTypeID string) string {
	stmt := fmt.Sprintf(`-- insert Location and location type
INSERT INTO public.location_types
	(location_type_id, name, "display_name", resource_path, updated_at, created_at)
	VALUES	('%s','org','%s', '%s', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO public.locations
	(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
	VALUES	('%s', '%s','%s',NULL, NULL, NULL, '%s', now(), now(),'%s') ON CONFLICT DO NOTHING;`,
		locationTypeID, name, resourcePath, locationID, name, locationTypeID, resourcePath, locationID)
	return stmt
}

func generateSchoolScript(resourcePath, name, tenantID string) string {
	iresourcePath, _ := strconv.Atoi(resourcePath)
	stmt := fmt.Sprintf(`-- insert School and organization_auths
INSERT INTO public.organization_auths
	(organization_id, auth_project_id, auth_tenant_id)
	VALUES(%d, 'staging-manabie-online', '%s') ON CONFLICT DO NOTHING;
	
INSERT INTO schools ( school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at, resource_path)
	VALUES (%d, '%s', 'COUNTRY_JP', 1, 1, null, false, now(), now(), false, null, null, '%s') ON CONFLICT DO NOTHING;`,
		iresourcePath, tenantID, iresourcePath, name, resourcePath)
	return stmt
}

func generatePermissionScript(resourcePath string) string {
	stmt := ` -- insert permission
INSERT INTO permission
	(permission_id, permission_name, created_at, updated_at, resource_path)
VALUES`

	for name, id := range ListPermissionIDs() {
		stmt = fmt.Sprintf("%s\n ('%s', '%s', now(), now(), '%s'),", stmt, id, name, resourcePath)
	}

	stmt = stmt[:len(stmt)-1] + ";"
	return stmt
}

func generateRoleScript(resourcePath string) string {
	stmt := ` -- insert role
INSERT INTO role
	(role_id, role_name, created_at, updated_at, resource_path)
VALUES`
	for name, id := range ListRoleIDs() {
		stmt = fmt.Sprintf("%s\n ('%s', '%s', now(), now(), '%s'),", stmt, id, name, resourcePath)
	}
	stmt = stmt[:len(stmt)-1] + ";"

	return stmt
}

func generatePermissionRoleScript(resourcePath string) string {
	stmt := `-- insert permission_role
INSERT INTO permission_role
	(permission_id, role_id, created_at, updated_at, resource_path)
VALUES`
	listRoleIDs := ListRoleIDs()
	listPermissionIDs := ListPermissionIDs()

	for roleName, listPermissions := range listRolePermission {
		for _, permissionID := range listPermissions {
			stmt = fmt.Sprintf("%s\n ('%s', '%s', now(), now(), '%s'),", stmt, listPermissionIDs[permissionID], listRoleIDs[roleName], resourcePath)
		}
	}
	stmt = stmt[:len(stmt)-1] + ";"

	return stmt
}

var randPool = sync.Pool{
	New: func() interface{} {
		// Note that this implementation to create the entropy:
		//	t := time.Now()
		//	return ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		// is not correct if there have *extremely* multiple concurrent calls.
		return ulid.Monotonic(rand.Reader, 0)
	},
}

func ULID(t time.Time) string {
	entropy := randPool.Get().(*ulid.MonotonicEntropy)
	defer randPool.Put(entropy)

	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// ULIDNow returns a new ULID.
func ULIDNow() string {
	return ULID(time.Now())
}

func ListPermissionIDs() map[string]string {
	permissionIDs := map[string]string{}
	if len(listPermission) == 0 {
		for _, permissions := range listRolePermission {
			for _, permission := range permissions {
				permissionIDs[permission] = ULIDNow()
			}
		}
		listPermission = permissionIDs
	} else {
		permissionIDs = listPermission
	}

	return permissionIDs
}

func ListRoleIDs() map[string]string {
	roleIDs := map[string]string{}
	if len(listRole) == 0 {
		for role := range listRolePermission {
			roleIDs[role] = ULIDNow()
		}
		listRole = roleIDs
	} else {
		roleIDs = listRole
	}
	return roleIDs
}

var listRolePermission = map[string][]string{
	"School Admin": {
		"master.location.read",
		"payment.invoice.write",
		"payment.invoice.read",
		"payment.payment.write",
		"user.student.read",
		"user.student.write",
		"user.studentpaymentdetail.read",
		"user.studentpaymentdetail.write",
		"user.student_course.write",
		"entryexit.student_entryexit_records.write",
		"entryexit.student_entryexit_records.read",
		"entryexit.student_qr.write",
		"entryexit.student_qr.read",
		"user.parent.read",
		"user.parent.write",
		"user.staff.read",
		"user.staff.write",
		"lesson.lessonmember.read",
		"lesson.lessonmember.write",
		"lesson.reallocation.read",
		"lesson.reallocation.write",
		"virtualclassroom.room_state.read",
		"virtualclassroom.room_state.write",
		"user.usergroup.read",
		"user.usergroup.write",
		"user.usergroupmember.write",
		"payment.payment.read",
		"communication.conversation.read",
		"communication.conversation.write",
		"communication.notification.read",
		"communication.notification.write",
		"payment.student_payment_detail.read",
		"payment.order.write",
		"payment.order.read",
		"payment.bill_item.read",
		"payment.bill_item.write",
		"payment.student_product.read",
		"payment.student_product.write",
		"payment.student_payment_detail.write",
		"payment.billing_address.read",
		"payment.billing_address.write",
		"payment.bank_account.read",
		"timesheet.timesheet.read",
		"timesheet.timesheet.write",
		"master.location.write",
		"master.course.read",
		"master.course.write",
		"master.class.read",
		"master.class.write",
		"communication.notification.owner",
		"payment.bank_account.write",
		"lesson.lesson.write",
		"lesson.lesson.read",
		"user.user.read",
		"user.user.write",
		"user.student_enrollment_status_history.read",
		"lesson.report.review",
	},
	"Teacher": {
		"communication.conversation.read",
		"entryexit.student_entryexit_records.read",
		"lesson.report.review",
		"user.parent.write",
		"master.course.read",
		"lesson.lessonmember.read",
		"user.usergroup.read",
		"communication.notification.read",
		"user.parent.read",
		"lesson.lessonmember.write",
		"user.user.write",
		"entryexit.student_qr.read",
		"user.staff.read",
		"master.class.read",
		"lesson.reallocation.read",
		"lesson.reallocation.write",
		"user.student.read",
		"virtualclassroom.room_state.write",
		"user.user.read",
		"master.location.read",
		"lesson.lesson.write",
		"lesson.lesson.read",
		"entryexit.student_entryexit_records.write",
		"virtualclassroom.room_state.read",
		"communication.conversation.write",
		"communication.notification.owner",
	},
	"Student": {
		"user.usergroup.read",
		"lesson.lesson.write",
		"master.location.read",
		"lesson.lesson.read",
		"communication.notification.read",
		"master.course.read",
		"user.student.write",
		"user.student.read",
		"communication.conversation.read",
		"user.user.read",
		"entryexit.student_qr.write",
		"user.user.write",
		"communication.conversation.write",
		"entryexit.student_qr.read",
		"master.class.read",
		"entryexit.student_entryexit_records.read",
	},
	"Parent": {
		"user.user.read",
		"user.parent.read",
		"lesson.lesson.read",
		"entryexit.student_entryexit_records.read",
		"master.class.read",
		"entryexit.student_qr.read",
		"master.course.read",
		"master.location.read",
		"user.usergroup.read",
		"user.user.write",
		"communication.conversation.read",
		"communication.conversation.write",
		"user.parent.write",
		"user.student.read",
		"communication.notification.read",
	},
	"HQ Staff": {
		"user.staff.write",
		"user.student_course.write",
		"user.studentpaymentdetail.write",
		"user.studentpaymentdetail.read",
		"user.usergroup.read",
		"user.usergroup.write",
		"user.student.write",
		"user.usergroupmember.write",
		"payment.billing_address.write",
		"user.user.write",
		"payment.payment.read",
		"payment.bank_account.write",
		"user.student.read",
		"payment.payment.write",
		"payment.bill_item.write",
		"payment.order.read",
		"payment.bill_item.read",
		"payment.bank_account.read",
		"payment.invoice.read",
		"payment.invoice.write",
		"communication.notification.read",
		"payment.student_product.write",
		"payment.student_payment_detail.write",
		"payment.order.write",
		"communication.notification.write",
		"payment.student_payment_detail.read",
		"user.parent.write",
		"master.course.write",
		"user.staff.read",
		"user.parent.read",
		"payment.billing_address.read",
		"lesson.lessonmember.read",
		"master.course.read",
		"master.class.read",
		"lesson.lessonmember.write",
		"lesson.report.review",
		"entryexit.student_qr.read",
		"timesheet.timesheet.write",
		"entryexit.student_qr.write",
		"lesson.reallocation.read",
		"timesheet.timesheet.read",
		"entryexit.student_entryexit_records.read",
		"master.class.write",
		"lesson.reallocation.write",
		"entryexit.student_entryexit_records.write",
		"communication.notification.owner",
		"virtualclassroom.room_state.read",
		"master.location.read",
		"payment.student_product.read",
		"lesson.lesson.read",
		"lesson.lesson.write",
		"user.user.read",
		"virtualclassroom.room_state.write",
	},
	"Centre Lead": {
		"payment.billing_address.read",
		"user.parent.read",
		"payment.bank_account.read",
		"lesson.lessonmember.read",
		"payment.student_payment_detail.read",
		"user.user.write",
		"master.class.read",
		"entryexit.student_qr.read",
		"user.student.read",
		"lesson.lessonmember.write",
		"virtualclassroom.room_state.write",
		"master.location.read",
		"payment.student_product.read",
		"entryexit.student_qr.write",
		"payment.student_product.write",
		"master.course.read",
		"lesson.reallocation.read",
		"user.parent.write",
		"entryexit.student_entryexit_records.read",
		"user.usergroup.read",
		"user.user.read",
		"lesson.report.review",
		"entryexit.student_entryexit_records.write",
		"user.student.write",
		"lesson.lesson.read",
		"payment.order.read",
		"user.staff.read",
		"virtualclassroom.room_state.read",
		"payment.bill_item.read",
	},
	"Teacher Lead": {
		"master.course.read",
		"lesson.lesson.read",
		"lesson.lessonmember.write",
		"virtualclassroom.room_state.read",
		"virtualclassroom.room_state.write",
		"user.student.read",
		"user.parent.read",
		"user.staff.read",
		"user.user.read",
		"user.user.write",
		"master.location.read",
		"user.usergroup.read",
		"master.class.read",
		"user.parent.write",
		"lesson.report.review",
	},
	"Centre Manager": {
		"payment.bank_account.read",
		"payment.student_product.write",
		"communication.notification.read",
		"payment.billing_address.write",
		"user.usergroup.read",
		"virtualclassroom.room_state.write",
		"lesson.report.review",
		"virtualclassroom.room_state.read",
		"lesson.reallocation.write",
		"lesson.reallocation.read",
		"timesheet.timesheet.read",
		"timesheet.timesheet.write",
		"lesson.lessonmember.write",
		"lesson.lessonmember.read",
		"master.course.read",
		"user.staff.read",
		"payment.billing_address.read",
		"user.parent.write",
		"user.parent.read",
		"master.class.read",
		"entryexit.student_qr.read",
		"entryexit.student_qr.write",
		"entryexit.student_entryexit_records.read",
		"master.class.write",
		"entryexit.student_entryexit_records.write",
		"lesson.lesson.read",
		"communication.notification.owner",
		"lesson.lesson.write",
		"user.student_course.write",
		"user.user.read",
		"user.student.write",
		"user.student.read",
		"user.user.write",
		"payment.student_payment_detail.write",
		"payment.invoice.read",
		"payment.student_payment_detail.read",
		"master.location.read",
		"payment.bill_item.read",
		"payment.bill_item.write",
		"payment.order.read",
		"payment.order.write",
		"payment.bank_account.write",
		"payment.student_product.read",
	},
	"Centre Staff": {
		"master.class.write",
		"lesson.lesson.read",
		"payment.bill_item.write",
		"communication.notification.owner",
		"lesson.lesson.write",
		"user.student_course.write",
		"user.student.write",
		"user.user.read",
		"user.student.read",
		"communication.notification.read",
		"payment.student_product.write",
		"user.user.write",
		"payment.billing_address.write",
		"payment.student_payment_detail.write",
		"user.usergroup.read",
		"payment.order.read",
		"virtualclassroom.room_state.write",
		"payment.invoice.read",
		"virtualclassroom.room_state.read",
		"payment.student_payment_detail.read",
		"lesson.reallocation.write",
		"master.location.read",
		"lesson.reallocation.read",
		"lesson.lessonmember.write",
		"lesson.lessonmember.read",
		"user.staff.read",
		"master.course.read",
		"payment.bank_account.read",
		"payment.billing_address.read",
		"user.parent.write",
		"payment.bank_account.write",
		"user.parent.read",
		"master.class.read",
		"lesson.report.review",
		"entryexit.student_qr.write",
		"payment.order.write",
		"entryexit.student_entryexit_records.read",
		"payment.bill_item.read",
		"entryexit.student_entryexit_records.write",
	},
	"OpenAPI": {
		"user.user.write",
		"user.usergroup.read",
		"user.user.read",
		"user.student.read",
		"user.student.write",
		"master.location.read",
		"user.usergroupmember.write",
	},
	"PaymentScheduleJob": {
		"payment.bill_item.write",
		"payment.bill_item.read",
		"payment.student_product.write",
		"payment.order.read",
		"payment.order.write",
		"payment.student_product.read",
	},
	"UsermgmtScheduleJob": {
		"user.usergroupmember.write",
		"user.usergroup.write",
		"user.student.read",
		"user.usergroup.read",
		"user.student.write",
		"user.user.write",
		"user.parent.read",
		"user.user.read",
		"user.parent.write",
		"user.staff.write",
		"user.staff.read",
		"user.student_enrollment_status_history.read",
		"master.location.read",
	},
	"NotificationScheduleJob": {
		"master.location.read",
		"user.user.read",
		"communication.notification.write",
		"master.course.read",
		"master.class.read",
		"communication.notification.read",
		"user.student.read",
	},
}

var listgroupName = []string{"Teacher", "School Admin", "Student", "Parent", "HQ Staff", "Centre Lead", "Teacher Lead", "Centre Manager", "Centre Staff", "OpenAPI", "PaymentScheduleJob", "UsermgmtScheduleJob", "NotificationScheduleJob"}
