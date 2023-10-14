ALTER TABLE public.e2e_instances
    ADD COLUMN IF NOT EXISTS on_trunk BOOLEAN DEFAULT false;

ALTER TABLE public.e2e_steps
    ADD COLUMN IF NOT EXISTS on_trunk BOOLEAN DEFAULT false;

ALTER TABLE public.e2e_scenarios
    ADD COLUMN IF NOT EXISTS on_trunk BOOLEAN DEFAULT false;

ALTER TABLE public.e2e_features
    ADD COLUMN IF NOT EXISTS on_trunk BOOLEAN DEFAULT false;

UPDATE public.e2e_instances SET on_trunk = true
  WHERE ((flavor ->> 'fe_ref' = 'develop' OR flavor ->> 'fe_ref' = '') AND
    (flavor ->> 'me_ref' = 'develop' OR flavor ->> 'me_ref' = '') AND
      flavor ->> 'eibanam_ref' = 'develop' ) OR
    (flavor ->> 'fe_ref' ilike 'release/202%' AND flavor ->> 'me_ref' ilike 'release/202%'
      AND flavor ->> 'eibanam_ref' ilike 'release/202%');
