update students set resource_path = school_id;

update teachers set resource_path = school_ids[1];

update school_admins set resource_path = school_id;

update parents set resource_path =school_id::text;

update users set resource_path = school_id::text from students  where user_id = student_id; 

update users set resource_path = school_ids[1]::text from teachers  where user_id = teacher_id;

update users set resource_path = school_id from school_admins where user_id = school_admin_id ;

update users set resource_path = school_id from parents where user_id = parent_id ;

update classes set resource_path = school_id::text;

update class_members cm set resource_path = school_id::text from classes c where cm.class_id = c.class_id;

CREATE INDEX IF NOT exists users_resource_path_idx on users using btree(resource_path);
CREATE INDEX IF NOT exists students_resource_path_idx on students using btree(resource_path);
CREATE INDEX IF NOT exists teachers_resource_path_idx on teachers using btree(resource_path);
CREATE INDEX IF NOT exists school_admins_resource_path_idx on school_admins using btree(resource_path);
CREATE INDEX IF NOT exists parents_resource_path_idx on parents using btree(resource_path);
