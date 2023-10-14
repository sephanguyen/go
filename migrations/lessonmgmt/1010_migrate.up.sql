-- public.partner_dynamic_form_field_values definition
CREATE TABLE public.partner_dynamic_form_field_values (
	dynamic_form_field_value_id text NOT NULL,
	field_id text NOT NULL,
	lesson_report_detail_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	value_type text NULL,
	string_value text NULL,
	bool_value bool NULL,
	string_array_value _text NULL,
	int_array_value _int4 NULL,
	string_set_value _text NULL,
	int_set_value _int4 NULL,
	field_render_guide jsonb NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	int_value int4 NULL,
	CONSTRAINT partner_dynamic_form_field_values_pk PRIMARY KEY (dynamic_form_field_value_id),
	CONSTRAINT unique__lesson_report_detail_id__field_id UNIQUE (lesson_report_detail_id, field_id)
);

CREATE POLICY rls_partner_dynamic_form_field_values ON "partner_dynamic_form_field_values" USING (permission_check(resource_path, 'partner_dynamic_form_field_values')) WITH CHECK (permission_check(resource_path, 'partner_dynamic_form_field_values'));
CREATE POLICY rls_partner_dynamic_form_field_values_restrictive ON "partner_dynamic_form_field_values" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'partner_dynamic_form_field_values')) with check (permission_check(resource_path, 'partner_dynamic_form_field_values'));

ALTER TABLE "partner_dynamic_form_field_values" ENABLE ROW LEVEL security;
ALTER TABLE "partner_dynamic_form_field_values" FORCE ROW LEVEL security;


-- public.partner_form_configs definition
CREATE TABLE public.partner_form_configs (
	form_config_id text NOT NULL,
	partner_id int4 NOT NULL,
	feature_name text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	form_config_data jsonb NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT partner_form_configs_pk PRIMARY KEY (form_config_id)
);

CREATE POLICY rls_partner_form_configs ON "partner_form_configs" USING (permission_check(resource_path, 'partner_form_configs')) WITH CHECK (permission_check(resource_path, 'partner_form_configs'));
CREATE POLICY rls_partner_form_configs_restrictive ON "partner_form_configs" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'partner_form_configs')) with check (permission_check(resource_path, 'partner_form_configs'));

ALTER TABLE "partner_form_configs" ENABLE ROW LEVEL security;
ALTER TABLE "partner_form_configs" FORCE ROW LEVEL security;