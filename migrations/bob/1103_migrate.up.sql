ALTER TABLE partner_dynamic_form_field_values
    ADD COLUMN IF NOT EXISTS "int_value" INTEGER,
    DROP COLUMN "int_vale";