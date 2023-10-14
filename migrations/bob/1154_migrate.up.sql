update schools 
set resource_path = school_id::text
where resource_path is null or length(resource_path) = 0;

update students 
set resource_path = school_id::text
where resource_path is null or length(resource_path) = 0;

update teachers 
set resource_path = school_ids[1]::text
where resource_path is null or length(resource_path) = 0;

update school_admins 
set resource_path = school_id::text
where resource_path is null or length(resource_path) = 0;

update parents 
set resource_path = school_id::text
where resource_path is null or length(resource_path) = 0;

update users u
set resource_path = s.school_id::text 
from students s
where u.user_id = s.student_id
    and (u.resource_path is null or length(u.resource_path) = 0); 

update users u
set resource_path = t.school_ids[1]::text
from teachers t
where u.user_id = t.teacher_id
    and (u.resource_path is null or length(u.resource_path) = 0); 

update users u
set resource_path = sa.school_id::text 
from school_admins sa
where u.user_id = sa.school_admin_id
    and (u.resource_path is null or length(u.resource_path) = 0); 

update users u
set resource_path = p.school_id::text 
from parents p
where u.user_id = p.parent_id
    and (u.resource_path is null or length(u.resource_path) = 0); 

update users_groups ug 
set resource_path = u.resource_path::text 
from users u 
where ug.user_id = u.user_id 
    and (ug.resource_path is null or length(ug.resource_path) = 0); 

update groups g 
set resource_path = ug.resource_path::text
from users_groups ug 
where g.group_id = ug.group_id
	and(g.resource_path is null or length(g.resource_path) = 0);

update cities c 
set resource_path = s.resource_path::text
from schools s 
where c.city_id = s.city_id
	and(c.resource_path is null or length(c.resource_path) = 0);


CREATE INDEX IF NOT exists users_resource_path_idx on users using btree(resource_path);
CREATE INDEX IF NOT exists students_resource_path_idx on students using btree(resource_path);
CREATE INDEX IF NOT exists teachers_resource_path_idx on teachers using btree(resource_path);
CREATE INDEX IF NOT exists school_admins_resource_path_idx on school_admins using btree(resource_path);
CREATE INDEX IF NOT exists parents_resource_path_idx on parents using btree(resource_path); 
