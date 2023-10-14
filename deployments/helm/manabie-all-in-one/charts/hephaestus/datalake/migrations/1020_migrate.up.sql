CREATE SCHEMA IF NOT EXISTS mastermgmt;

CREATE TABLE IF NOT EXISTS mastermgmt.grade (
	grade_id text NOT NULL,
	"name" text NOT NULL,
	is_archived bool NOT NULL DEFAULT false,
	partner_internal_id varchar(50) NOT NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text,
	"sequence" int4 NULL,
	remarks text NULL,
	CONSTRAINT grade_pk PRIMARY KEY (grade_id)
);


CREATE TABLE IF NOT EXISTS mastermgmt.academic_year (
	academic_year_id text NOT NULL,
	"name" text NOT NULL,
	start_date date NOT NULL,
	end_date date NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text,
	CONSTRAINT pk__academic_year PRIMARY KEY (academic_year_id)
);


CREATE TABLE IF NOT EXISTS bob.locations (
    location_id text NOT NULL,
    name text,
    location_type text,
    parent_location_id text,
    partner_internal_id text,
    partner_internal_parent_id text,
    updated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz,
    is_archived boolean NOT NULL DEFAULT false,
    access_path text,
    resource_path text,
    CONSTRAINT location_pk PRIMARY KEY (location_id)
);

CREATE TABLE IF NOT EXISTS bob.location_types (
	location_type_id text NOT NULL,
	"name" text NOT NULL,
	display_name text NULL,
	parent_name text NULL,
	parent_location_type_id text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text,
	is_archived bool NOT NULL DEFAULT false,
	CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id)
);

CREATE TABLE IF NOT EXISTS bob.subject (
    subject_id TEXT NOT NULL PRIMARY KEY,
    name text NOT NULL,
    display_name text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text
);

CREATE TABLE IF NOT EXISTS bob.class (
    class_id TEXT NOT NULL,
    name TEXT NOT NULL,
    course_id TEXT NOT NULL,
    school_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
    CONSTRAINT pk__class PRIMARY KEY (class_id)
);

CREATE TABLE IF NOT EXISTS bob.class_member (
    class_member_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT,
    CONSTRAINT pk__class_member PRIMARY KEY (class_member_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
mastermgmt.grade,
mastermgmt.academic_year,
bob.locations,
bob.location_types,
bob.class,
bob.class_member,
bob.subject;
