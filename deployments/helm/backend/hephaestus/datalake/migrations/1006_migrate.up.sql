-- bob."role" definition

CREATE TABLE IF NOT EXISTS bob."role" (
	role_id text NOT NULL,
	role_name text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	is_system bool NULL DEFAULT false,
	CONSTRAINT role__pk PRIMARY KEY (role_id, resource_path)
);

-- bob.granted_role definition

CREATE TABLE IF NOT EXISTS bob.granted_role (
	granted_role_id text NOT NULL,
	user_group_id text NOT NULL,
	role_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT granted_role_granted_role_id_key UNIQUE (granted_role_id),
	CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id)
);

-- bob.granted_role_access_path definition

CREATE TABLE IF NOT EXISTS bob.granted_role_access_path (
	granted_role_id text NOT NULL,
	location_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id)
);

-- bob."permission" definition

CREATE TABLE IF NOT EXISTS bob."permission" (
	permission_id text NOT NULL,
	permission_name text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);

-- bob.permission_role definition

CREATE TABLE IF NOT EXISTS bob.permission_role (
	permission_id text NOT NULL,
	role_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT permission_role__pk PRIMARY KEY (permission_id, role_id, resource_path)
);

-- bob.granted_permission definition

CREATE TABLE IF NOT EXISTS bob.granted_permission (
	user_group_id text NOT NULL,
	user_group_name text NOT NULL,
	role_id text NOT NULL,
	role_name text NOT NULL,
	permission_id text NOT NULL,
	permission_name text NOT NULL,
	location_id text NOT NULL,
	resource_path text NOT NULL,
	CONSTRAINT granted_permission__pk PRIMARY KEY (user_group_id, role_id, permission_id, location_id)
);

-- bob.user_group definition

CREATE TABLE IF NOT EXISTS bob.user_group (
	user_group_id text NOT NULL,
	user_group_name text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	org_location_id text NULL,
	is_system bool NULL DEFAULT false,
	CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);

-- bob.user_group_member definition

CREATE TABLE IF NOT EXISTS bob.user_group_member (
	user_id text NOT NULL,
	user_group_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL,
	CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);

-- bob.student_enrollment_status_history definition

CREATE TABLE IF NOT EXISTS bob.student_enrollment_status_history (
	student_id text NOT NULL,
	location_id text NOT NULL,
	enrollment_status text NOT NULL,
	start_date timestamptz NOT NULL,
	end_date timestamptz NULL,
	"comment" text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NULL,
	order_id text NULL,
	order_sequence_number int4 NULL,
	CONSTRAINT pk__student_enrollment_status_history PRIMARY KEY (student_id, location_id, enrollment_status, start_date),
	CONSTRAINT student_enrollment_status_his_student_id_location_id_enroll_key UNIQUE (student_id, location_id, enrollment_status, start_date, end_date),
	CONSTRAINT students_enrollment_status_check CHECK ((enrollment_status = ANY ('{STUDENT_ENROLLMENT_STATUS_POTENTIAL,STUDENT_ENROLLMENT_STATUS_ENROLLED,STUDENT_ENROLLMENT_STATUS_TEMPORARY,STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,STUDENT_ENROLLMENT_STATUS_WITHDRAWN,STUDENT_ENROLLMENT_STATUS_GRADUATED,STUDENT_ENROLLMENT_STATUS_LOA}'::text[])))
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.role;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.granted_role;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.granted_role_access_path;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.permission;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.permission_role;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.granted_permission;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.user_group;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.user_group_member;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.student_enrollment_status_history;
