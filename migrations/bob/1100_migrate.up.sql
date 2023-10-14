-- Change primary key of lesson_report_details table
ALTER TABLE lesson_report_details DROP CONSTRAINT lesson_report_details_pk;
ALTER TABLE lesson_report_details
    ADD COLUMN IF NOT EXISTS lesson_report_detail_id TEXT;
ALTER TABLE lesson_report_details
    ADD CONSTRAINT lesson_report_details_pk PRIMARY KEY (lesson_report_detail_id);
ALTER TABLE lesson_report_details
    ADD CONSTRAINT unique__lesson_report_id__student_id UNIQUE (lesson_report_id, student_id);

CREATE TABLE IF NOT EXISTS public.partner_dynamic_form_field_values (
    "dynamic_form_field_value_id" TEXT NOT NULL,
    "field_id" TEXT NOT NULL,
    "lesson_report_detail_id" TEXT NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "value_type" TEXT NOT NULL,
    "int_vale" INTEGER,
    "string_value" TEXT,
    "bool_value" BOOLEAN,
    "string_array_value" TEXT[],
    "int_array_value" INTEGER[],
    "string_set_value" TEXT[],
    "int_set_value" INTEGER[],
    "field_render_guide" JSONB,
    "resource_path" TEXT,
    CONSTRAINT lesson_report_detail_id_fk FOREIGN KEY (lesson_report_detail_id) REFERENCES public.lesson_report_details(lesson_report_detail_id),
    CONSTRAINT unique__lesson_report_detail_id__field_id UNIQUE (lesson_report_detail_id, field_id),
    CONSTRAINT partner_dynamic_form_field_values_pk PRIMARY KEY (dynamic_form_field_value_id)
);

CREATE TABLE IF NOT EXISTS public.partner_form_configs (
    "form_config_id" TEXT NOT NULL,
    "partner_id" INTEGER NOT NULL,
    "feature_name" TEXT NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "form_config_data" JSONB,
    "resource_path" TEXT,
    CONSTRAINT partner_id_fk FOREIGN KEY (partner_id) REFERENCES public.schools(school_id),
    CONSTRAINT unique__partner_id__feature_name UNIQUE (partner_id, feature_name),
    CONSTRAINT partner_form_configs_pk PRIMARY KEY (form_config_id)
);

ALTER TABLE lesson_reports
    ADD COLUMN IF NOT EXISTS form_config_id TEXT;
ALTER TABLE lesson_reports
    ADD CONSTRAINT form_config_id_fk FOREIGN KEY (form_config_id) REFERENCES public.partner_form_configs(form_config_id);
