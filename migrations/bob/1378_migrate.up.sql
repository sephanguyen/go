DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='debezium_publication') THEN
      CREATE PUBLICATION debezium_publication;
   END IF;
END
$do$;

ALTER PUBLICATION debezium_publication SET TABLE 
public.dbz_signals,
public.locations,
public.location_types,
public.organizations,
public.grade,
public.users,
public.user_access_paths,
public.students,
public.courses,
public.school_admins,
public.student_parents,
public.staff,
public.lessons_courses,
public.lessons,
public.lessons_teachers,
public.lesson_members,
public.course_access_paths,
public.granted_role,
public.role,
public.user_group_member,
public.user_group,
public.groups,
public.users_groups,
public.permission,
public.permission_role,
public.granted_role_access_path,
public.student_qr,
public.student_entryexit_records,
public.prefecture,
public.user_tag,
public.tagged_user,
public.granted_permission;
