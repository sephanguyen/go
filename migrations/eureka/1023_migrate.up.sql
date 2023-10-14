CREATE TABLE IF NOT EXISTS "student_latest_submissions" (
  study_plan_item_id text NOT NULL,
  assignment_id text NOT NULL,
  student_id text NOT NULL,
  student_submission_id text NOT NULL,
  submission_content jsonb,
  check_list jsonb,
  status text,
  note text,
  editor_id text,
  student_submission_grade_id text,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,

  CONSTRAINT student_latest_submissions_pk PRIMARY KEY (student_id, study_plan_item_id, assignment_id),
  CONSTRAINT student_latest_submission_assigment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id),
  CONSTRAINT student_latest_submission_study_plan_item_fk FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS student_latest_submissions_submission_id ON student_latest_submissions (student_submission_id DESC);

INSERT INTO student_latest_submissions
SELECT study_plan_item_id, assignment_id, student_id, student_submission_id, submission_content, check_list, status,
      note, editor_id, student_submission_grade_id, created_at, updated_at, deleted_at
FROM (
	SELECT ((array_agg(student_submissions.* ORDER BY student_submissions.student_submission_id DESC))[1]).*
	FROM student_submissions
	GROUP BY student_id, study_plan_item_id, assignment_id
) AS submissions
ON CONFLICT DO NOTHING;
