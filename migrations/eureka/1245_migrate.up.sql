UPDATE lo_submission_answer lsa
SET
  submitted_keys_answer = sh.submitted_keys_answer,
  correct_keys_answer = sh.correct_keys_answer
FROM
  get_submission_history() AS sh
  JOIN quizzes q ON q.external_id = sh.quiz_id
WHERE
  lsa.deleted_at IS NULL
  AND sh.shuffled_quiz_set_id = lsa.shuffled_quiz_set_id
  AND q.kind = 'QUIZ_TYPE_ORD'
  AND q.deleted_at IS NULL;