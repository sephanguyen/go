CREATE OR REPLACE FUNCTION public.find_teacher_by_school_id_v2(school_id integer) RETURNS SETOF public.users
    LANGUAGE sql STABLE
    AS $$
    select u.* from  teachers t join users u on u.user_id = t.teacher_id where 
    u.deleted_at is null and t.deleted_at is null and 
    case WHEN school_id != 0 then t.school_ids @> ARRAY[school_id]
 	else 1 = 1
    end
$$;
