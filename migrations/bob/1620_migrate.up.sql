ALTER TABLE public.student_parents DROP CONSTRAINT IF EXISTS student_parents__student_id__fk;
ALTER TABLE public.student_parents DROP CONSTRAINT IF EXISTS student_parents__parent_id__fk;

-- delete records having the invalid reference id;
WITH cte AS (
	SELECT sp.student_id FROM public.student_parents sp 
    LEFT JOIN public.students s ON s.student_id = sp.student_id
    WHERE s.student_id IS NULL
)
DELETE FROM public.student_parents sp WHERE sp.student_id IN (SELECT * FROM cte);
WITH cte AS (
	SELECT sp.parent_id FROM public.student_parents sp 
    LEFT JOIN public.parents s ON s.parent_id = sp.parent_id
    WHERE s.parent_id IS NULL
)
DELETE FROM public.student_parents sp WHERE sp.parent_id IN (SELECT * FROM cte);

-- add the constraints
ALTER TABLE public.student_parents 
    ADD CONSTRAINT student_parents__student_id__fk FOREIGN KEY(student_id) REFERENCES public.students(student_id),
    ADD CONSTRAINT student_parents__parent_id__fk FOREIGN KEY(parent_id) REFERENCES public.parents(parent_id);
    
