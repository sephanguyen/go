ALTER TABLE public."configuration" ADD config_value_type text NOT NULL DEFAULT 'string'::text;
ALTER TABLE public."external_configuration" ADD config_value_type text NOT NULL DEFAULT 'string'::text;
