DROP FUNCTION IF EXISTS public.get_first_parent_from_student_ids(student_ids text[]);

CREATE OR REPLACE FUNCTION public.get_first_parent_from_student_ids(student_ids TEXT[])
    RETURNS SETOF public.student_parents
    LANGUAGE sql STABLE
AS
$$
SELECT sp.student_id, sp.parent_id, sp.created_at, sp.updated_at, sp.deleted_at, sp.relationship, sp.resource_path
FROM (SELECT sp1.*,
             row_number() over (PARTITION BY sp1.student_id ORDER BY sp1.created_at ASC) as seqnum
      FROM student_parents as sp1 WHERE sp1.deleted_at IS NULL
     ) sp
WHERE sp.seqnum = 1
  AND sp.student_id = ANY (student_ids);
$$;
