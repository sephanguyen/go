CREATE TABLE IF NOT EXISTS public.student_submission_grades (
	student_submission_grade_id text NOT NULL,
	student_submission_id text NOT NULL,
	grade float4 NULL,
	grade_content JSONB,
	grader_id text NOT NULL,
	grader_comment text NULL,

    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,

	CONSTRAINT student_submission_grades_pk PRIMARY KEY (student_submission_grade_id),
	CONSTRAINT student_submission_grades_fk FOREIGN KEY (student_submission_id) REFERENCES public.student_submissions(student_submission_id)
);
