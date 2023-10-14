INSERT INTO assessment_submission
(id,
 session_id,
 assessment_id,
 student_id,
 grading_status,
 max_score,
 graded_score,
 created_at,
 updated_at,
 resource_path,
 completed_at)
select generate_ulid() as id
     , session_id
     , assessment_id
     , user_id         as student_id
     , 'RETURNED'      as grading_status
     , 0               as max_score
     , 0               as graded_score
     , updated_at      as created_at -- because session status was updated to COMPLETED
     , updated_at
     , resource_path
     , updated_at      as completed_at
from assessment_session
where status = 'COMPLETED'
ON CONFLICT ON CONSTRAINT session_id_un DO NOTHING; -- ignore existing submission
