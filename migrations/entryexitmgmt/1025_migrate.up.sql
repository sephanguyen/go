-- Set the latest sequence of student_entryexit_records to the max entryexit_id
SELECT SETVAL('public."student_entryexit_records_id_seq"', COALESCE(MAX(entryexit_id), 1)) FROM public."student_entryexit_records";

-- Set the latest sequence of student_qr to the max qr_id
SELECT SETVAL('public."student_qr_id_seq"', COALESCE(MAX(qr_id), 1)) FROM public."student_qr";