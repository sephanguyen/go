CREATE TABLE IF NOT EXISTS fatima.product_grade (
    product_id TEXT NOT NULL,
    grade_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT product_grade_pk PRIMARY KEY (product_id,grade_id)
);

CREATE TABLE IF NOT EXISTS fatima.product_price (
    product_price_id INTEGER NOT NULL,
    product_id TEXT NOT NULL,
    billing_schedule_period_id TEXT,
    quantity INTEGER,
    price numeric(12,2) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT product_price_pk PRIMARY KEY (product_price_id)
);

CREATE TABLE IF NOT EXISTS fatima.package_course_fee (
    package_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    fee_id TEXT NOT NULL,
    quantity INTEGER,
    available_from timestamp with time zone,
    available_until timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT package_course_fee_pk PRIMARY KEY (package_id,course_id,fee_id)
);

CREATE TABLE IF NOT EXISTS fatima.package_course_material (
    package_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    material_id TEXT NOT NULL,
    available_from timestamp with time zone,
    available_until timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT package_course_material_pk PRIMARY KEY (package_id,course_id,material_id)
);

CREATE TABLE IF NOT EXISTS fatima.product_accounting_category (
    product_id TEXT NOT NULL,
    accounting_category_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT product_accounting_category_pk PRIMARY KEY (product_id,accounting_category_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
fatima.product_grade,
fatima.product_price,
fatima.package_course_fee,
fatima.package_course_material,
fatima.product_accounting_category;
