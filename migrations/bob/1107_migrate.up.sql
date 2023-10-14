ALTER TABLE lesson_reports
DROP COLUMN IF EXISTS school_id;

ALTER TABLE lesson_report_details
DROP COLUMN IF EXISTS course_id;

ALTER TABLE partner_dynamic_form_field_values
    ALTER COLUMN value_type DROP NOT NULL;

ALTER TABLE partner_form_configs DROP CONSTRAINT unique__partner_id__feature_name;
