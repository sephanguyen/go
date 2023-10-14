update public.e2e_steps e2estep
  set on_trunk = e2es.on_trunk 
  from public.e2e_scenarios e2es
  where e2es.on_trunk = true and e2es.scenario_id = e2estep.scenario_id ;