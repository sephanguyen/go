UPDATE
  student_submissions ss
SET
  complete_date = COALESCE(spi.completed_at, ss.created_at)
FROM
  study_plan_items spi
WHERE
  ss.complete_date IS NULL
  AND ss.study_plan_item_id = spi.study_plan_item_id
  AND spi.completed_at IS NOT NULL;