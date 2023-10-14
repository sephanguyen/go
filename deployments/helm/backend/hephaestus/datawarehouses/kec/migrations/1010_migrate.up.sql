-- bob.partner_dynamic_form_field_values definition

CREATE TABLE IF NOT EXISTS bob.partner_dynamic_form_field_values_public_info (
	dynamic_form_field_value_id text NOT NULL,
	field_id text NOT NULL,
	lesson_report_detail_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	value_type text NULL,
	string_value text NULL,
	int_value int4 NULL,
	CONSTRAINT partner_dynamic_form_field_values_pk PRIMARY KEY (dynamic_form_field_value_id),
	CONSTRAINT unique__lesson_report_detail_id__field_id UNIQUE (lesson_report_detail_id, field_id)
);
