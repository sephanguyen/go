ALTER TABLE public.e2e_scenarios
ADD COLUMN IF NOT EXISTS feature_path text;

CREATE TABLE IF NOT EXISTS public.e2e_scenario_severity (
    scenario_name text NOT NULL,
  	feature_path text NOT NULL,
  	feature_name text NOT NULL,
    keyword text NOT NULL,
	severity_tags text NOT NULL,
  	PRIMARY KEY(feature_path, scenario_name)
);

ALTER TABLE public.e2e_scenarios
drop constraint if exists scenario_severity_fk;

ALTER TABLE public.e2e_scenarios
ADD CONSTRAINT scenario_severity_fk
FOREIGN KEY (feature_path, name)
REFERENCES public.e2e_scenario_severity(feature_path, scenario_name);
