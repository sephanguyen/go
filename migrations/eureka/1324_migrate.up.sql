CREATE TABLE IF NOT EXISTS "cerebry_classes" (
    "id" text NOT NULL,
    "class_code" text NOT NULL,
    "class_name" text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT cerebry_classes_pk PRIMARY KEY (id),
    CONSTRAINT cerebry_classes_name_un UNIQUE (class_code)
);

/* set RLS */
CREATE POLICY rls_cerebry_classes ON "cerebry_classes"
USING (permission_check(resource_path, 'cerebry_classes')) WITH CHECK (permission_check(resource_path, 'cerebry_classes'));
CREATE POLICY rls_cerebry_classes_restrictive ON "cerebry_classes" AS RESTRICTIVE
USING (permission_check(resource_path, 'cerebry_classes')) WITH CHECK (permission_check(resource_path, 'cerebry_classes'));

ALTER TABLE "cerebry_classes" ENABLE ROW LEVEL security;
ALTER TABLE "cerebry_classes" FORCE ROW LEVEL security;

ALTER TABLE "course_students"
    ADD COLUMN vendor_synced_at timestamptz NULL;

ALTER TABLE "courses"
    ADD COLUMN is_adaptive boolean default false,
    ADD COLUMN vendor_id text NULL;

ALTER TABLE "courses"
    ADD CONSTRAINT fk_cerebry_class_id FOREIGN KEY (vendor_id) REFERENCES cerebry_classes (id);
