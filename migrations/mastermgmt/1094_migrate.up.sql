CREATE TABLE IF NOT EXISTS public.configuration_group (
    configuration_group_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value BOOLEAN DEFAULT false,
    description TEXT,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    partner_resource_path text NOT NULL,
    CONSTRAINT configuration_group_pk PRIMARY KEY (configuration_group_id)
);


CREATE TABLE IF NOT EXISTS public.configuration_group_map (
    configuration_group_id TEXT NOT NULL,
    configuration_value_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    partner_resource_path text NOT NULL,
    CONSTRAINT configuration_group_map_pk PRIMARY KEY(configuration_group_id, configuration_value_id)
);


ALTER TABLE ONLY public.configuration_key ADD COLUMN IF NOT EXISTS description text;
