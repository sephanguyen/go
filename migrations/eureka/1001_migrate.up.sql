CREATE TABLE IF NOT EXISTS "study_plans" (
  "study_plan_id" text,
  "master_study_plan_id" text,
  "name" text,
  "study_plan_type" text,
  "school_id" integer,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT study_plans_pk PRIMARY KEY (study_plan_id)
);

CREATE TABLE IF NOT EXISTS "study_plan_items" (
  "study_plan_item_id" text,
  "study_plan_id" text,
  "available_from" timestamp with time zone,
  "start_date" timestamp with time zone,
  "end_date" timestamp with time zone,
  "deleted_at" timestamp with time zone,
  "available_to" timestamp with time zone,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  CONSTRAINT study_plan_items_pk PRIMARY KEY (study_plan_item_id)
);

CREATE TABLE IF NOT EXISTS "student_study_plans" (
  "study_plan_id" text,
  "student_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT student_study_plans_pk PRIMARY KEY (study_plan_id, student_id)
);

CREATE TABLE IF NOT EXISTS "course_study_plans" (
  "course_id" text,
  "study_plan_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT course_study_plans_pk PRIMARY KEY (course_id, study_plan_id)
);

CREATE TABLE IF NOT EXISTS "class_study_plans" (
  "class_id" integer,
  "study_plan_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT class_study_plans_pk PRIMARY KEY (class_id, study_plan_id)
);

CREATE TABLE IF NOT EXISTS "lesson_study_plan_items" (
  "lesson_id" text,
  "study_plan_item_id" text,
  "course_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT lesson_study_plan_items_pk PRIMARY KEY (lesson_id, study_plan_item_id)
);

CREATE TABLE IF NOT EXISTS "assignments" (
  "assignment_id" text,
  "content" jsonb,
  "attachment" text[],
  "settings" jsonb,
  /*[
      {
        allow_re-submission: bool
        allow_late_submission: bool
        require_attachment: bool
        require_assignment_note: bool
        require_video_submission: bool
      }
  ]*/
  "check_list" jsonb,
  /*[
      {
        "string": bool
      }
  ]*/
  "name" text NOT NULL,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  "max_grade" integer,
  "status" text,
  "instruction" text,
  "type" text,
  CONSTRAINT assignments_pk PRIMARY KEY (assignment_id)
);

CREATE TABLE IF NOT EXISTS "assignment_study_plan_items" (
  "assignment_id" text,
  "study_plan_item_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT assignment_study_plan_items_pk PRIMARY KEY (study_plan_item_id, assignment_id)
);

CREATE TABLE IF NOT EXISTS "student_submissions" (
  "student_submission_id" text NOT NULL,
  "study_plan_item_id" text NOT NULL,
  "assignment_id" text NOT NULL,
  "student_id" text NOT NULL,
  "submission_content" jsonb,
  /* {
    submit_media_id: string,
    attachment_media_id: string
  }*/
  "check_list" jsonb,
  "status" text,
  "note" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,

  CONSTRAINT student_submissions_pk PRIMARY KEY (student_submission_id),
	CONSTRAINT student_submission_assigment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id),
	CONSTRAINT student_submission_study_plan_item_fk FOREIGN KEY (study_plan_item_id) REFERENCES public.study_plan_items(study_plan_item_id)
);

CREATE TABLE IF NOT EXISTS "course_classes" (
  "course_id" text NOT NULL,
  "class_id" text NOT NULL,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT course_classes_pk PRIMARY KEY (course_id,class_id)
);

CREATE TABLE IF NOT EXISTS "course_students" (
  "course_id" text,
  "student_id" text,
  "created_at" timestamp with time zone NOT NULL,
  "updated_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT course_student_pk PRIMARY KEY (course_id,student_id)
);

CREATE TABLE IF NOT EXISTS "class_students" (
  "student_id" text,
  "class_id" text,
  "updated_at" timestamp with time zone NOT NULL,
  "created_at" timestamp with time zone NOT NULL,
  "deleted_at" timestamp with time zone,
  CONSTRAINT class_students_pk PRIMARY KEY (student_id,class_id)
);

ALTER TABLE "study_plan_items" ADD FOREIGN KEY ("study_plan_id") REFERENCES "study_plans" ("study_plan_id");

ALTER TABLE "study_plans" ADD FOREIGN KEY ("master_study_plan_id") REFERENCES "study_plans" ("study_plan_id");

ALTER TABLE "assignment_study_plan_items" ADD FOREIGN KEY ("study_plan_item_id") REFERENCES "study_plan_items" ("study_plan_item_id");

ALTER TABLE "assignment_study_plan_items" ADD FOREIGN KEY ("assignment_id") REFERENCES "assignments" ("assignment_id");

ALTER TABLE "course_study_plans" ADD FOREIGN KEY ("study_plan_id") REFERENCES "study_plans" ("study_plan_id");

ALTER TABLE "student_study_plans" ADD FOREIGN KEY ("study_plan_id") REFERENCES "study_plans" ("study_plan_id");

ALTER TABLE "class_study_plans" ADD FOREIGN KEY ("study_plan_id") REFERENCES "study_plans" ("study_plan_id");
