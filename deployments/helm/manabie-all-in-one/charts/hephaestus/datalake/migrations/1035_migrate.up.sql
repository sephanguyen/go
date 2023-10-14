CREATE TABLE IF NOT EXISTS fatima.file (
	file_id text NOT NULL,
	file_name text NOT NULL,
	file_type text NOT NULL,
	download_link text NOT NULL,
	updated_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	created_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text,
	CONSTRAINT file_pk PRIMARY KEY (file_id)
);

ALTER TABLE fatima.order_item ALTER updated_at DROP DEFAULT;
ALTER TABLE fatima.order_item alter COLUMN updated_at DROP NOT NULl;

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
fatima.file;

