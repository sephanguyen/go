
UPDATE public.e2e_features e2ef 
  SET on_trunk  = e2ei.on_trunk 
  FROM public.e2e_instances e2ei 
  WHERE e2ei.on_trunk = true AND e2ef.instance_id  = e2ei.instance_id;
