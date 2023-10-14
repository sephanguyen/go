CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.partner_form_configs (
    form_config_id TEXT NOT NULL,
    partner_id INTEGER NOT NULL,
    feature_name TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    form_config_data JSONB NULL,
    resource_path TEXT NOT NULL,
    CONSTRAINT pk__partner_form_configs PRIMARY KEY (form_config_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.partner_form_configs;
