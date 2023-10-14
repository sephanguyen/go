CREATE TABLE IF NOT EXISTS public.leaving_reason (
	leaving_reason_id text NOT NULL,
	leaving_reason_name text NOT NULL,
	leaving_reason_type text NOT NULL,
	leaving_reason_remark text NULL,
	leaving_reason_is_archived bool NOT NULL DEFAULT false,
	leaving_reason_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    leaving_reason_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    leaving_reason_deleted_at timestamp with time zone,
	CONSTRAINT leaving_reason_pk PRIMARY KEY (leaving_reason_id)
);

CREATE TABLE IF NOT EXISTS public.order_item_course (
	order_id text NOT NULL,
	package_id text NULL,
	course_id text NOT NULL,
	course_name text NOT NULL,
	course_slot int4 NULL,
	course_slot_per_week int4 NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	order_item_course_id text NOT NULL,
	CONSTRAINT order_item_course_id_pk PRIMARY KEY (order_item_course_id)
);

CREATE TABLE IF NOT EXISTS public.package_quantity_type_mapping (
	package_type text NOT NULL,
	quantity_type text NOT NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	CONSTRAINT package_quantity_type_mapping_pk PRIMARY KEY (package_type)
);

CREATE TABLE IF NOT EXISTS public.bill_item_course (
	bill_item_sequence_number int4 NOT NULL,
	course_id text NOT NULL,
	course_name text NOT NULL,
	course_weight int4 NULL,
	course_slot int4 NULL,
	course_slot_per_week int4 NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	CONSTRAINT bill_item_course_pk PRIMARY KEY (bill_item_sequence_number, course_id)
);


CREATE TABLE IF NOT EXISTS public.student_product (
	student_product_id text NOT NULL,
	student_id text NOT NULL,
	product_id text NOT NULL,
	upcoming_billing_date timestamptz NULL,
	start_date timestamptz NULL,
	end_date timestamptz NULL,
	product_status text NOT NULL,
	approval_status text NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	location_id text NOT NULL,
	updated_from_student_product_id text NULL,
	updated_to_student_product_id text NULL,
	student_product_label text NULL,
	is_unique bool NULL DEFAULT false,
	root_student_product_id text NULL,
	is_associated bool NULL DEFAULT false,
	version_number int4 NOT NULL DEFAULT 0,
	CONSTRAINT student_product_pk PRIMARY KEY (student_product_id)
);

CREATE TABLE IF NOT EXISTS public.order_action_log (
	order_action_log_id int4 NOT NULL,
	staff_id text NOT NULL,
	order_id text NOT NULL,
	"action" text NULL,
	"comment" text NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	CONSTRAINT order_action_log_pk PRIMARY KEY (order_action_log_id)
);

CREATE TABLE IF NOT EXISTS public.product (
	product_id text NOT NULL,
	"name" text NOT NULL,
	product_type text NOT NULL,
	tax_id text NULL,
	available_from timestamptz NOT NULL,
	available_until timestamptz NOT NULL,
	remarks text NULL,
	custom_billing_period timestamptz NULL,
	billing_schedule_id text NULL,
	disable_pro_rating_flag bool NOT NULL DEFAULT false,
	is_archived bool NOT NULL DEFAULT false,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	is_unique bool NULL DEFAULT false,
	product_tag text NULL,
	product_partner_id text NULL,
	CONSTRAINT product_pk PRIMARY KEY (product_id)
);

CREATE TABLE IF NOT EXISTS public.accounting_category (
	accounting_category_id text NOT NULL,
	"name" text NOT NULL,
	remarks text NULL,
	is_archived bool NOT NULL DEFAULT false,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	CONSTRAINT accounting_category_pk PRIMARY KEY (accounting_category_id)
);

CREATE TABLE IF NOT EXISTS public.product_setting (
	product_id text NOT NULL,
	is_enrollment_required bool NULL DEFAULT false,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	is_pausable bool NULL DEFAULT true,
	is_added_to_enrollment_by_default bool NULL DEFAULT false,
	is_operation_fee bool NULL DEFAULT false,
	CONSTRAINT product_settings_pk PRIMARY KEY (product_id)
);

CREATE TABLE IF NOT EXISTS public.package_course (
	package_id text NOT NULL,
	course_id text NOT NULL,
	mandatory_flag bool NOT NULL DEFAULT false,
	course_weight int4 NOT NULL DEFAULT 1,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	max_slots_per_course int4 NOT NULL DEFAULT 1,
	CONSTRAINT package_course_pk PRIMARY KEY (package_id, course_id)
);

CREATE TABLE IF NOT EXISTS public.product_location (
	product_id text NOT NULL,
	location_id text NOT NULL,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
	CONSTRAINT product_location_pk PRIMARY KEY (product_id, location_id)
);

ALTER PUBLICATION kec_publication ADD TABLE 
public.leaving_reason,
public.order_item_course,
public.package_quantity_type_mapping,
public.bill_item_course,
public.student_product,
public.order_action_log,
public.product,
public.accounting_category,
public.product_setting,
public.package_course,
public.product_location;
