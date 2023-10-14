CREATE INDEX IF NOT EXISTS e2e_instances_created_at__idx ON public.e2e_instances (created_at DESC NULLS FIRST);

CREATE INDEX IF NOT EXISTS e2e_features_started_at__idx ON public.e2e_features (started_at DESC NULLS FIRST);
CREATE INDEX IF NOT EXISTS e2e_features_name__idx ON public.e2e_features (name);


CREATE INDEX IF NOT EXISTS e2e_scenarios_started_at__idx ON public.e2e_scenarios (started_at ASC NULLS FIRST);
CREATE INDEX IF NOT EXISTS e2e_scenarios_name__idx ON public.e2e_scenarios (name);

CREATE INDEX IF NOT EXISTS e2e_steps_name__idx ON public.e2e_steps (name);
CREATE INDEX IF NOT EXISTS e2e_steps_index__idx ON public.e2e_steps (index ASC NULLS FIRST);
