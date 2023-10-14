CREATE TABLE IF NOT EXISTS package_discount_course_mapping (
    package_id TEXT NOT NULL,
    course_combination_ids TEXT NOT NULL,
    discount_tag_id TEXT NOT NULL,
    is_archived BOOLEAN DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT package_discount_course_mapping__pk PRIMARY KEY (package_id,discount_tag_id),
    CONSTRAINT package_discount_course_mapping_discount_tag_fk FOREIGN KEY (discount_tag_id) REFERENCES "discount_tag"(discount_tag_id),
    CONSTRAINT package_discount_course_mapping_package_fk FOREIGN KEY (package_id) REFERENCES "package"(package_id),
    CONSTRAINT unique__package_id_course_combination_ids_discount_tag_id UNIQUE (package_id, discount_tag_id,course_combination_ids)
    
);

CREATE POLICY rls_package_discount_course_mapping ON "package_discount_course_mapping"
USING (permission_check(resource_path, 'package_discount_course_mapping'))
WITH CHECK (permission_check(resource_path, 'package_discount_course_mapping'));

CREATE POLICY rls_package_discount_course_mapping_restrictive ON "package_discount_course_mapping" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'package_discount_course_mapping'))
WITH CHECK (permission_check(resource_path, 'package_discount_course_mapping'));

ALTER TABLE "package_discount_course_mapping" ENABLE ROW LEVEL security;
ALTER TABLE "package_discount_course_mapping" FORCE ROW LEVEL security;
