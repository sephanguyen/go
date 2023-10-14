INSERT INTO course_type
  (course_type_id, name, created_at, updated_at, resource_path)
VALUES 
  ('29GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483629'),
  ('29GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483629'),

  ('30GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483630'),
  ('30GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483630'),
  
  ('31GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483631'),
  ('31GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483631'),
  
  ('32GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483632'),
  ('32GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483632'),
  
  ('33GD2BBHB50W6CTHS9WAWTXAA0', 'Regular', now(), now(), '-2147483633'),
  ('33GD2BBHB50W6CTHS9WAWTXAA1', 'Seasonal', now(), now(), '-2147483633') ON CONFLICT DO NOTHING;
