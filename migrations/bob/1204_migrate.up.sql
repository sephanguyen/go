INSERT INTO public.location_types
(location_type_id, name, "display_name", resource_path, updated_at, created_at)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8M1','org','E2E Tokyo', '-2147483639', now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.locations
(location_id, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at,access_path)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q8N1', 'E2E Tokyo','01FR4M51XJY9E77GSN4QZ1Q8M1',NULL, NULL, NULL, '-2147483639', now(), now(),'01FR4M51XJY9E77GSN4QZ1Q8N1') ON CONFLICT DO NOTHING;

