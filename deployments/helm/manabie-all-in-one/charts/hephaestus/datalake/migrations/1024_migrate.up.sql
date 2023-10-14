CREATE TABLE IF NOT EXISTS fatima.fee (
    fee_id TEXT NOT NULL,
    fee_type TEXT NOT NULL,
    created_at timestamp with time zone NULL,
    updated_at timestamp with time zone NULL,
    deleted_at timestamp with time zone NULL,
    resource_path text NOT NULL,
    CONSTRAINT pk_fee_id PRIMARY KEY (fee_id)
);

CREATE TABLE IF NOT EXISTS fatima.package (
    package_id TEXT NOT NULL,
    package_type TEXT NOT NULL,
    max_slot INTEGER NOT NULL,
    package_start_date timestamp with time zone,
    package_end_date timestamp with time zone,
    created_at timestamp with time zone NULL,
    updated_at timestamp with time zone NULL,
    deleted_at timestamp with time zone NULL,
    resource_path text NOT NULL,
    CONSTRAINT pk_package_id PRIMARY KEY (package_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE
fatima.fee,
fatima.package;
