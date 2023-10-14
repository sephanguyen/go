-- parent
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Parent'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
		user.parent.read,
		user.parent.write,
		user.student.read,
		user.user.read
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- student
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Student'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
		user.student.read,
		user.student.write,
		user.user.read
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- teacher
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Teacher'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
		user.parent.read,
		user.parent.write,
		user.student.read,
		user.user.read
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- teacher lead
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Teacher Lead'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
		user.parent.read,
		user.parent.write,
		user.student.read,
		user.user.read
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- school admin
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'School Admin'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
    user.parent.read,
    user.parent.write,
    user.staff.read,
    user.staff.write,
    user.student_course.write,
    user.studentpaymentdetail.read,
    user.studentpaymentdetail.write,
    user.student.read,
    user.student.write,
    user.usergroupmember.write,
    user.usergroup.read,
    user.usergroup.write,
    user.user.read,
    user.user.write
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- hq staff
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'HQ Staff'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
    user.parent.read,
    user.parent.write,
    user.staff.read,
    user.staff.write,
    user.student_course.write,
    user.studentpaymentdetail.read,
    user.studentpaymentdetail.write,
    user.student.read,
    user.student.write,
    user.usergroupmember.write,
    user.usergroup.read,
    user.usergroup.write,
    user.user.read,
    user.user.write
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- centre manager
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Centre Manager'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
    user.parent.read,
    user.parent.write,
    user.staff.read,
    user.student_course.write,
    user.student.read,
    user.student.write,
    user.usergroup.read,
    user.user.read,
    user.user.write
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- centre staff
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Centre Staff'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
    user.parent.read,
    user.parent.write,
    user.staff.read,
    user.student_course.write,
    user.student.read,
    user.student.write,
    user.usergroup.read,
    user.user.read,
    user.user.write
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;

-- centre lead
INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'Centre Lead'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
    user.parent.read,
    user.parent.write,
    user.staff.read,
    user.student.read,
    user.student.write,
    user.usergroup.read,
    user.user.read,
    user.user.write
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;
