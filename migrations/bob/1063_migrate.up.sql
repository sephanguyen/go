CREATE TABLE IF NOT EXISTS student_parents
(
	student_id TEXT,
	parent_id TEXT,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT student_parents_pk PRIMARY KEY(student_id, parent_id)
);

