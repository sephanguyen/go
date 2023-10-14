CREATE TABLE IF NOT EXISTS bob.role_public_info (
    role_id text NOT NULL,
	role_name text NOT NULL,
	role_created_at timestamptz NOT NULL,
	role_updated_at timestamptz NOT NULL,
	role_deleted_at timestamptz NULL,

    granted_role_id text NOT NULL,
	user_group_id text NOT NULL,
	granted_role_created_at timestamptz NOT NULL,
	granted_role_updated_at timestamptz NOT NULL,
	granted_role_deleted_at timestamptz NULL,

	location_id text NOT NULL,
	granted_role_access_path_created_at timestamptz NOT NULL,
	granted_role_access_path_updated_at timestamptz NOT NULL,
	granted_role_access_path_deleted_at timestamptz NULL,

    CONSTRAINT pk__role PRIMARY KEY (role_id)
);

CREATE TABLE IF NOT EXISTS bob.permission_public_info (
    permission_id text NOT NULL,
	permission_permission_name text NOT NULL,
	permission_created_at timestamptz NOT NULL,
	permission_updated_at timestamptz NOT NULL,
	permission_deleted_at timestamptz NULL,

    permission_role_role_id text NOT NULL,
	permission_role_created_at timestamptz NOT NULL,
	permission_role_updated_at timestamptz NOT NULL,
	permission_role_deleted_at timestamptz NULL,

    user_group_id text NOT NULL,
	user_group_name text NOT NULL,
	granted_permission_role_id text NOT NULL,
	role_name text NOT NULL,
	granted_permission_permission_name text NOT NULL,
	location_id text NOT NULL,

    CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);

CREATE TABLE IF NOT EXISTS bob.user_group_public_info (
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

CREATE TABLE IF NOT EXISTS bob.user_group_member_public_info (
    user_id text NOT NULL,
	user_group_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);

CREATE TABLE IF NOT EXISTS bob.student_enrollment_status_history_public_info (
    student_id text NOT NULL,
	location_id text NOT NULL,
	enrollment_status text NOT NULL,
	start_date timestamptz NOT NULL,
	end_date timestamptz NULL,
	"comment" text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	order_id text NULL,
	order_sequence_number int4 NULL,

    CONSTRAINT pk__student_enrollment_status_history PRIMARY KEY (student_id,location_id,enrollment_status,start_date)
);
