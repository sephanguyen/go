ALTER TABLE organizations ALTER COLUMN organization_id TYPE text;

INSERT INTO public.organizations
(organization_id, tenant_id, "name", resource_path)
VALUES	('-2147483648','-2147483648','Manabie School', '-2147483648'),
		('-2147483647','-2147483647','JPREP School', '-2147483647'),
		('-2147483646','-2147483646','Synersia School','-2147483646'),
		('-2147483645','-2147483645','Renseikai School','-2147483645'),
		('-2147483644','-2147483644','End-to-end School','-2147483644'),
		('-2147483643','-2147483643','GA School','-2147483643'),
		('-2147483642','-2147483642','KEC School','-2147483642'),
		('-2147483641','-2147483641','AIC School','-2147483641'),
		('-2147483640','-2147483640','NSG School','-2147483640');