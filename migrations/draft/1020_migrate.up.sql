update public.e2e_scenarios e2es
  set on_trunk  = e2ef.on_trunk 
  from public.e2e_features  e2ef
  where e2ef.on_trunk = true and e2es.feature_id = e2ef.feature_id; 