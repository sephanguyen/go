CREATE TABLE IF NOT EXISTS public.location_types (
	location_type_id TEXT NOT NULL PRIMARY KEY,
	name text NOT NULL,
	display_name text,
	parent_name text,
	parent_location_type_id text,
	updated_at timestamp with time zone NOT NULL,
	created_at timestamp with time zone NOT NULL,
	deleted_at timestamp with time zone,
	resource_path text DEFAULT autofillresourcepath(),
	CONSTRAINT location_type_id_fk FOREIGN KEY (parent_location_type_id) REFERENCES location_types(location_type_id),
	CONSTRAINT unique__location_type_name_resource_path UNIQUE (name, resource_path)
);

CREATE POLICY rls_location_types ON "location_types" using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));

ALTER TABLE "location_types" ENABLE ROW LEVEL security;
ALTER TABLE "location_types" FORCE ROW LEVEL security;

/* rename column parent_id of locations */
ALTER TABLE locations RENAME COLUMN parent_id TO parent_location_id;

/* Set FK for parent_location_id */
ALTER TABLE public.locations DROP CONSTRAINT IF EXISTS location_id_fk;

/* Migrate location_type to null */
UPDATE public.locations SET location_type = NULL;

ALTER TABLE IF EXISTS locations ADD CONSTRAINT location_id_fk FOREIGN KEY (parent_location_id) REFERENCES locations(location_id);

/* Change location_type default value to null */
ALTER TABLE public.locations ALTER location_type SET DEFAULT NULL;

/* Gennerate default value */

INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q9M1','org','Manabie Org', '-2147483648', now(), now() ),
		('01FR4M51XJY9E77GSN4QZ1Q9M2','org','JPREP Org', '-2147483647', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M3','org','Synersia Org','-2147483646', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M4','org','Renseikai Org','-2147483645', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M5','org','End-to-end Org','-2147483644', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M6','org','GA Org','-2147483643', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M7','org','KEC Org','-2147483642', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M8','org','AIC Org','-2147483641', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9M9','org','NSG Org','-2147483640', now(), now());

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q9N1', 'Manabie','01FR4M51XJY9E77GSN4QZ1Q9M1','1', NULL, NULL, '-2147483648', now(), now() ),
		('01FR4M51XJY9E77GSN4QZ1Q9N2', 'JPREP','01FR4M51XJY9E77GSN4QZ1Q9M2','1', NULL, NULL, '-2147483647', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N3', 'Synersia','01FR4M51XJY9E77GSN4QZ1Q9M3','1', NULL, NULL, '-2147483646', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N4', 'Renseikai','01FR4M51XJY9E77GSN4QZ1Q9M4','1', NULL, NULL, '-2147483645', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N5', 'End-to-end','01FR4M51XJY9E77GSN4QZ1Q9M5','1', NULL, NULL, '-2147483644', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N6', 'GA','01FR4M51XJY9E77GSN4QZ1Q9M6','1', NULL, NULL, '-2147483643', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N7', 'KEC','01FR4M51XJY9E77GSN4QZ1Q9M7','1', NULL, NULL, '-2147483642', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N8', 'AIC','01FR4M51XJY9E77GSN4QZ1Q9M8','1', NULL, NULL, '-2147483641', now(), now()),
		('01FR4M51XJY9E77GSN4QZ1Q9N9', 'NSG','01FR4M51XJY9E77GSN4QZ1Q9M9','1', NULL, NULL, '-2147483640', now(), now());

/* Set FK for location_type */
ALTER TABLE IF EXISTS locations ADD CONSTRAINT location_type_fk FOREIGN KEY (location_type) REFERENCES location_types(location_type_id);