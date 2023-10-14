CREATE TABLE IF NOT EXISTS public.course_type (
  course_type_id TEXT NOT NULL,
  name TEXT NOT NULL,

  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  resource_path TEXT DEFAULT autofillresourcepath(),

  CONSTRAINT course_type__pk PRIMARY KEY (course_type_id)
);

CREATE POLICY rls_course_type ON "course_type"
USING (permission_check(resource_path, 'course_type'))
WITH CHECK (permission_check(resource_path, 'course_type'));

CREATE POLICY rls_course_type_restrictive ON "course_type" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'course_type'))
with check (permission_check(resource_path, 'course_type'));


ALTER TABLE IF EXISTS public.course_type ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.course_type FORCE ROW LEVEL security;

ALTER TABLE "courses"
  ADD COLUMN IF NOT EXISTS course_type_id TEXT;

ALTER TABLE public.courses ADD CONSTRAINT fk__course__course_type_id FOREIGN KEY(course_type_id) REFERENCES course_type(course_type_id);

INSERT INTO course_type
  (course_type_id, name, created_at, updated_at, resource_path)
VALUES 
  ('34GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483634'),
  ('34GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483634'),

  ('35GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483635'),
  ('35GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483635'),
  
  ('37GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483637'),
  ('37GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483637'),
  
  ('38GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483638'),
  ('38GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483638'),
  
  ('39GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483639'),
  ('39GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483639'),
  
  ('40GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483640'),
  ('40GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483640'),
  
  ('41GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483641'),
  ('41GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483641'),
  
  ('42GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483642'),
  ('42GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483642'),
  
  ('43GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483643'),
  ('43GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483643'),
  
  ('44GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483644'),
  ('44GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483644'),
  
  ('45GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483645'),
  ('45GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483645'),
  
  ('46GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483646'),
  ('46GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483646'),
  
  ('47GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483647'),
  ('47GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483647'),
  
  ('48GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483648'),
  ('48GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483648');