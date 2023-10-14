-- Run this statement before for the update students_topics_completeness below
ALTER TABLE students_topics_completeness REPLICA IDENTITY FULL;

update users_groups ug set resource_path = u.resource_path::text
from users u
where ug.user_id = u.user_id
    and(ug.resource_path is null or length(ug.resource_path)=0);

update users_info_notifications uin set resource_path = u.resource_path::text
from users u
where uin.user_id = u.user_id
    and(uin.resource_path is null or length(uin.resource_path)=0);

update groups g set resource_path = ug.resource_path::text
from users_groups ug 
where g.group_id = ug.group_id
	and(g.resource_path is null or length(g.resource_path)=0);

update students_topics_completeness stc set resource_path = s.school_id::text
from students s 
where stc.student_id = s.student_id
	and(stc.resource_path is null or length(stc.resource_path)=0);

update students_learning_objectives_completeness sloc set resource_path = s.school_id::text
from students s 
where sloc.student_id = s.student_id
	and(sloc.resource_path is null or length(sloc.resource_path)=0);

update student_parents sp set resource_path = s.school_id::text
from students s 
where sp.student_id = s.student_id
	and(sp.resource_path is null or length(sp.resource_path)=0);

update students_learning_objectives_records slor set resource_path = s.school_id::text
from students s 
where slor.student_id = s.student_id
	and(slor.resource_path is null or length(slor.resource_path)=0);
