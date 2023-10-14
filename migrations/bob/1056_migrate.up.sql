CREATE TABLE IF NOT EXISTS public.students_learning_objectives_records (
	record_id text not null,
	lo_id text not null,
	study_plan_item_id text not null,
	student_id text not null,
	accuracy numeric(5,2),
	learning_time int,
	completed_at timestamp with time zone NOT NULL,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL,
	deleted_at timestamp with time zone,
	is_offline boolean,
	CONSTRAINT students_learning_objectives_records_pk PRIMARY KEY (record_id),
	CONSTRAINT students_learning_objectives_records_lo_id_fk FOREIGN KEY (lo_id) REFERENCES learning_objectives(lo_id),
	CONSTRAINT students_learning_objectives_records_student_id_fk FOREIGN KEY (student_id) REFERENCES students(student_id)
);
