CREATE TABLE IF NOT EXISTS package_discount_setting (
    package_id TEXT NOT NULL,
    min_slot_trigger INTEGER,
    max_slot_trigger INTEGER,
    discount_tag_id TEXT NOT NULL,
    is_archived BOOLEAN DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT package_discount_setting__pk PRIMARY KEY (package_id,discount_tag_id),
    CONSTRAINT package_discount_setting_discount_tag_fk FOREIGN KEY (discount_tag_id) REFERENCES "discount_tag"(discount_tag_id),
    CONSTRAINT package_discount_setting_package_fk FOREIGN KEY (package_id) REFERENCES "package"(package_id)
);

CREATE POLICY rls_package_discount_setting ON "package_discount_setting"
USING (permission_check(resource_path, 'package_discount_setting'))
WITH CHECK (permission_check(resource_path, 'package_discount_setting'));

CREATE POLICY rls_package_discount_setting_restrictive ON "package_discount_setting"
AS RESTRICTIVE TO public
USING (permission_check(resource_path, 'package_discount_setting'))
WITH CHECK (permission_check(resource_path, 'package_discount_setting'));

ALTER TABLE "package_discount_setting" ENABLE ROW LEVEL security;
ALTER TABLE "package_discount_setting" FORCE ROW LEVEL security;
