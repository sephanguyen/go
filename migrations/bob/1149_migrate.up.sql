ALTER TABLE "lesson_student_subscriptions"
DROP CONSTRAINT IF EXISTS lesson_student_subscriptions_pkey;

ALTER TABLE "lesson_student_subscriptions"
ADD CONSTRAINT lesson_student_subscriptions_pkey PRIMARY KEY (student_subscription_id);

ALTER TABLE "lesson_student_subscriptions"
ADD CONSTRAINT lesson_student_subscriptions_uniq UNIQUE (subscription_id, course_id, student_id);

-- update new data (start_at, end_at, updated_at) for all records
with lastest
AS (
    select l1.* from lesson_student_subscriptions l1 JOIN (
        SELECT course_id , student_id , max(created_at) max_created_at
        FROM lesson_student_subscriptions
        WHERE deleted_at is null
        GROUP BY course_id, student_id
    ) l2
    ON l1.course_id = l2.course_id AND l1.student_id = l2.student_id AND l1.created_at = l2.max_created_at
)
UPDATE lesson_student_subscriptions l
SET start_at = lastest.start_at, end_at = lastest.end_at, updated_at = lastest.updated_at
FROM lastest
WHERE l.course_id = lastest.course_id and l.student_id = lastest.student_id;

-- delete all duplicated except oldest records
with oldest
AS (
    select l1.* from lesson_student_subscriptions l1 JOIN (
        SELECT course_id , student_id , min(created_at) min_created_at
        FROM lesson_student_subscriptions
        WHERE deleted_at is null
        GROUP BY course_id, student_id
    ) l2
    ON l1.course_id = l2.course_id AND l1.student_id = l2.student_id AND l1.created_at = l2.min_created_at
)
UPDATE lesson_student_subscriptions l
SET deleted_at = now()
WHERE l.student_subscription_id not in (select student_subscription_id from oldest);
