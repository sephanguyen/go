CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.partner_form_configs_public_info (
    form_config_id TEXT NOT NULL,
    partner_id INTEGER NOT NULL,
    feature_name TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    form_config_data JSONB NULL,
    CONSTRAINT pk__partner_form_configs PRIMARY KEY (form_config_id)
);
