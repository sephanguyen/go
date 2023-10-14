ALTER TABLE ONLY public.invoice_action_log
  DROP CONSTRAINT IF EXISTS invoice_action_log_users_fk,
  ADD CONSTRAINT invoice_action_log_user_basic_info__fk FOREIGN KEY (user_id) REFERENCES user_basic_info (user_id);

ALTER TABLE ONLY public.billing_address
  DROP CONSTRAINT IF EXISTS billing_address_users_fk,
  ADD CONSTRAINT billing_address_user_basic_info__fk FOREIGN KEY (user_id) REFERENCES user_basic_info (user_id);


ALTER TABLE ONLY public.invoice_schedule 
  DROP CONSTRAINT IF EXISTS invoice_schedule_users_fk,
  ADD CONSTRAINT billing_address_user_basic_info__fk FOREIGN KEY (user_id) REFERENCES user_basic_info (user_id);

ALTER TABLE ONLY public.student_payment_detail_action_log 
  DROP CONSTRAINT IF EXISTS student_payment_detail_action_log__users__fk,
  ADD CONSTRAINT student_payment_detail_action_log_user_basic_info__fk FOREIGN KEY (user_id) REFERENCES user_basic_info (user_id);
