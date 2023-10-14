-- update value of sequence "student_event_logs_student_event_log_id_seq" with maximum value of column "student_event_log_id" from table student_event_logs
SELECT SETVAL('student_event_logs_student_event_log_id_seq', COALESCE((SELECT MAX(student_event_log_id) FROM student_event_logs), 1));

-- update value of sequence "student_learning_time_by_daily_learning_time_id_seq" with maximum value of column "learning_time_id" from table student_learning_time_by_daily
SELECT SETVAL('student_learning_time_by_daily_learning_time_id_seq', COALESCE((SELECT MAX(learning_time_id) FROM student_learning_time_by_daily), 1));
