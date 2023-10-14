CREATE OR REPLACE FUNCTION public.find_teacher_by_resource_path(resource_path_id varchar) RETURNS SETOF public.users
    LANGUAGE sql STABLE
    AS $$
    select u.*
    from staff s
    left join users u on s.staff_id = u.user_id
    where s.deleted_at is null and u.deleted_at is null
    and s.resource_path = resource_path_id
    and u.resource_path = resource_path_id
$$;
