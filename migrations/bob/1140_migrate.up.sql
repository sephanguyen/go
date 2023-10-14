ALTER TABLE "lesson_student_subscriptions"
    DROP CONSTRAINT IF EXISTS lesson_student_subscriptions_pkey;
ALTER TABLE "lesson_student_subscriptions"
    ADD CONSTRAINT lesson_student_subscriptions_pkey PRIMARY KEY (subscription_id, course_id, student_id);