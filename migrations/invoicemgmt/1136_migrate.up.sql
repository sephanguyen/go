ALTER TABLE ONLY public.invoice_schedule 
    DROP CONSTRAINT IF EXISTS billing_address_user_basic_info__fk,
    ADD CONSTRAINT invoice_schedule_user_basic_info__fk FOREIGN KEY (user_id) REFERENCES user_basic_info (user_id);
