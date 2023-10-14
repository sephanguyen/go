ALTER TABLE ONLY public.student_payment_detail_action_log
    ALTER COLUMN action_detail TYPE JSONB USING action_detail::jsonb;
