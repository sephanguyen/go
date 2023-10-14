update lessons l set resource_path = c.school_id::text
from courses c
where l.course_id = c.course_id
	and(l.resource_path is null or length(l.resource_path)=0);
	
update lesson_members lm set resource_path = l.resource_path
from lessons l
where l.lesson_id = lm.lesson_id
	and(lm.resource_path is null or length(lm.resource_path)=0);
	
update lessons_teachers lt set resource_path = l.resource_path
from lessons l
where l.lesson_id = lt.lesson_id 
	and(lt.resource_path is null or length(lt.resource_path)=0);