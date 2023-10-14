CREATE OR REPLACE FUNCTION public.get_first_parent_from_student_ids(student_ids TEXT[])
    RETURNS TABLE
            (
                student_id   text,
                parent_id    text,
                relationship text
            )
    LANGUAGE plpgsql
AS
$$
BEGIN
RETURN QUERY SELECT sp.student_id, sp.parent_id, sp.relationship
                 FROM (SELECT sp1.*,
                              row_number() over (PARTITION BY sp1.student_id ORDER BY sp1.created_at ASC ) as seqnum
                       FROM student_parents as sp1
                      ) sp
                 WHERE sp.seqnum = 1
                   AND sp.student_id = ANY (student_ids);
END;
$$;