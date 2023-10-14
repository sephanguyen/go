-- Add product_group_id into package_discount_setting table
-- Step 1: Add the new field
ALTER TABLE package_discount_setting
    ADD COLUMN product_group_id TEXT NOT NULL;

-- Step 2: Add the foreign key constraint
ALTER TABLE package_discount_setting
    ADD CONSTRAINT package_discount_setting_product_group_fk
        FOREIGN KEY (product_group_id) REFERENCES product_group(product_group_id);

-- Step 3: Remove the existing primary key constraint
ALTER TABLE package_discount_setting
DROP CONSTRAINT package_discount_setting__pk;

-- Step 4: Add the new composite primary key
ALTER TABLE package_discount_setting
    ADD CONSTRAINT package_discount_setting__pk
        PRIMARY KEY (package_id, discount_tag_id, product_group_id);

-- Add product_group_id into package_discount_course_mapping table
-- Step 1: Add the new field
ALTER TABLE package_discount_course_mapping
    ADD COLUMN product_group_id TEXT NOT NULL;

-- Step 2: Add the foreign key constraint
ALTER TABLE package_discount_course_mapping
    ADD CONSTRAINT package_discount_course_mapping_product_group_fk
        FOREIGN KEY (product_group_id) REFERENCES product_group(product_group_id);

-- Step 3: Remove the existing primary key constraint
ALTER TABLE package_discount_course_mapping
DROP CONSTRAINT package_discount_course_mapping__pk;

-- Step 4: Add the new composite primary key
ALTER TABLE package_discount_course_mapping
    ADD CONSTRAINT package_discount_course_mapping__pk
        PRIMARY KEY (package_id, discount_tag_id, product_group_id);

-- Step 5: Add the unique constraint
ALTER TABLE package_discount_course_mapping
    ADD CONSTRAINT unique__package_id_course_combination_ids_discount_tag_id_product_group_id
        UNIQUE (package_id, discount_tag_id, course_combination_ids, product_group_id);
