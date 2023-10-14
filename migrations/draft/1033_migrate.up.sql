ALTER TABLE public.e2e_scenarios
ADD COLUMN IF NOT EXISTS raw_name text;

ALTER TABLE public.e2e_scenarios
DROP CONSTRAINT IF EXISTS scenario_severity_fk;

ALTER TABLE public.e2e_scenarios
ADD CONSTRAINT scenario_severity_fk
FOREIGN KEY (feature_path, raw_name)
REFERENCES public.e2e_scenario_severity(feature_path, scenario_name);

ALTER TABLE public.e2e_scenario_severity
ADD COLUMN IF NOT EXISTS "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
ADD COLUMN IF NOT EXISTS "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc');