DROP VIEW IF EXISTS public.e2e_instances_feature_tags;

DROP VIEW IF EXISTS public.e2e_instances_squad_tags;

CREATE VIEW public.e2e_instances_feature_tags AS
    select distinct unnest(tags) as feature_tag from e2e_instances eei where deleted_at is null ORDER BY feature_tag ASC;

CREATE VIEW public.e2e_instances_squad_tags AS 
select distinct unnest(squad_tags) as squad_tag from e2e_instances eei where deleted_at is null ORDER BY squad_tag ASC;

