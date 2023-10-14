CREATE TABLE IF NOT EXISTS public.grade (
	grade_id text NOT NULL,
	"name" text NOT NULL,
	is_archived bool NOT NULL DEFAULT false,
	partner_internal_id varchar(50) NOT NULL,
	grade_updated_at timestamptz NOT NULL,
	grade_created_at timestamptz NOT NULL,
	grade_deleted_at timestamptz NULL,
	"sequence" int4 NULL,
	remarks text NULL,
	CONSTRAINT grade_pk PRIMARY KEY (grade_id)
);


CREATE TABLE IF NOT EXISTS public.academic_year (
	academic_year_id text NOT NULL,
	"name" text NOT NULL,
	start_date date NOT NULL,
	end_date date NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	CONSTRAINT pk__academic_year PRIMARY KEY (academic_year_id)
);


CREATE TABLE IF NOT EXISTS public.locations (
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
    CONSTRAINT location_pk PRIMARY KEY (location_id)
);

CREATE TABLE IF NOT EXISTS public.location_types (
	location_type_id text NOT NULL,
	"name" text NOT NULL,
	display_name text NULL,
	parent_name text NULL,
	parent_location_type_id text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	is_archived bool NOT NULL DEFAULT false,
	CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id)
);

CREATE TABLE IF NOT EXISTS public.subject (
    subject_id TEXT NOT NULL PRIMARY KEY,
    name text NOT NULL,
    display_name text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS public.class_class_members (
    class_member_id TEXT NOT NULL PRIMARY KEY,
    class_id text NOT NULL,
    student_id text NOT NULL,
    class_member_updated_at timestamp with time zone NOT NULL,
    class_member_created_at timestamp with time zone NOT NULL,
    class_member_deleted_at timestamp with time zone,
	start_date timestamp with time zone,
	end_date timestamp with time zone,
	name text NOT NULL,
	course_id text NOT NULL,
	school_id text NULL,
	location_id text NULL,
	class_updated_at timestamp with time zone NOT NULL,
    class_created_at timestamp with time zone NOT NULL,
    class_deleted_at timestamp with time zone
);

ALTER PUBLICATION kec_publication ADD TABLE 
public.grade,
public.academic_year,
public.locations,
public.location_types,
public.class_class_members,
public.subject;
