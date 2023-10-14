-- Lesson Report permission and role
-- Create permission: lesson.report.review for all partner except Manabie B2C Internal(-2147483634) & E2E Architecture(100000)
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GS7JJ7W4ATHFT6YT4M4CBK9A', 'lesson.report.review', now(), now(), '-2147483648'), -- Manabie / GA Test
  ('01GS7JJ7W4ATHFT6YT4M4CBK9B', 'lesson.report.review', now(), now(), '-2147483647'), -- JPREP
  ('01GS7JJ7W4ATHFT6YT4M4CBK9C', 'lesson.report.review', now(), now(), '-2147483646'), -- Synersia
  ('01GS7JJ7W4ATHFT6YT4M4CBK9D', 'lesson.report.review', now(), now(), '-2147483645'), -- Renseikai
  ('01GS7JJ7W4ATHFT6YT4M4CBK9E', 'lesson.report.review', now(), now(), '-2147483644'), -- E2E / GA Prod
  ('01GS7JJ7W4ATHFT6YT4M4CBK9F', 'lesson.report.review', now(), now(), '-2147483643'), -- GA UAT
  ('01GS7JJ7W4ATHFT6YT4M4CBK9G', 'lesson.report.review', now(), now(), '-2147483642'), -- KEC
  ('01GS7JJ7W4ATHFT6YT4M4CBK9H', 'lesson.report.review', now(), now(), '-2147483641'), -- AIC
  ('01GS7JJ7W4ATHFT6YT4M4CBK9I', 'lesson.report.review', now(), now(), '-2147483640'), -- NSG
  ('01GS7JJ7W4ATHFT6YT4M4CBK9J', 'lesson.report.review', now(), now(), '-2147483639'), -- E2E Tokyo
  ('01GS7JJ7W4ATHFT6YT4M4CBK9K', 'lesson.report.review', now(), now(), '-2147483638'), -- E2E HCM
  ('01GS7JJ7W4ATHFT6YT4M4CBK9L', 'lesson.report.review', now(), now(), '-2147483637'), -- Manabie Demo LMS
  ('01GS7JJ7W4ATHFT6YT4M4CBK9M', 'lesson.report.review', now(), now(), '-2147483635'), -- KEC Demo
  ('01GS7JJ7W4ATHFT6YT4M4CBK9N', 'lesson.report.review', now(), now(), '-2147483631'), -- Eishinkan
  ('01GS7JJ7W4ATHFT6YT4M4CBK9O', 'lesson.report.review', now(), now(), '-2147483630'), -- Withus Base
  ('01GS7JJ7W4ATHFT6YT4M4CBK9P', 'lesson.report.review', now(), now(), '-2147483629'), -- Withus HighSchool
  ('01GS7JJ7W4ATHFT6YT4M4CBK9Q', 'lesson.report.review', now(), now(), '-2147483628') -- Manabie Demo ERP
  ON CONFLICT DO NOTHING;