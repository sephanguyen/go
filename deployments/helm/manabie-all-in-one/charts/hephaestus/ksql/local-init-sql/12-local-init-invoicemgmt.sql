\connect invoicemgmt;

INSERT INTO public.invoice_adjustment (invoice_adjustment_id,invoice_id,description,amount,student_id,resource_path,created_at,updated_at,deleted_at) VALUES
    ('01H79W6JZZQA42KHR6S0WJ1VRJ','01GWHQV538V3AP0WVDFMZ0TJ2W','test-descript',2.00,'01GWHQY4BBRYZ7G0XDEQN22TN0','-2147483642',now(),now(),NULL) ON CONFLICT DO NOTHING;;

INSERT INTO public.invoice_schedule (invoice_schedule_id,invoice_date,status,resource_path,created_at,updated_at,user_id,scheduled_date) VALUES
    ('01H7AK56A2W5JTPCAVQ887TRXX',now(),'test-status-dwh','-2147483642',now(),now(),'01GX5MP57FS52F1MGDQ7SJ6CDS',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.invoice_schedule_history (invoice_schedule_history_id,invoice_schedule_id,number_of_failed_invoices,total_students,execution_start_date,execution_end_date,resource_path,created_at,updated_at) VALUES
    ('01H7AK9F5P1AQEBGRV8HZBX4CD','01H7AK56A2W5JTPCAVQ887TRXX',0,1,now(),now(),'-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.invoice_schedule_student (invoice_schedule_student_id,invoice_schedule_history_id,student_id,error_details,resource_path,created_at) VALUES
    ('01H7AKE2E13CQA5FKTWMBTJ6YK','01H7AK9F5P1AQEBGRV8HZBX4CD','01GWHQY4BBRYZ7G0XDEQN22TN0','test-error-dwh','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.invoice_action_log (invoice_action_id,invoice_id,user_id,action,action_detail,action_comment,resource_path,created_at,updated_at) VALUES
    ('01H7C104AE3FRDPZN9KQJ9D1ZR','01GWHQV538V3AP0WVDFMZ0TJ2W','01GX5MP57FS52F1MGDQ7SJ6CDS','test-action-dwh','test-action-detail-dwh','test-action-comment','-2147483642',now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.new_customer_code_history (new_customer_code_history_id,new_customer_code,student_id,bank_account_number,resource_path,created_at,updated_at) VALUES
    ('01H7C7433H2TTX8M5JJEGYGTPZ','1','01GWHQY4BBRYZ7G0XDEQN22TN0','test-bank-number-dwh','-2147483642',now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.billing_address (billing_address_id,user_id,student_payment_detail_id,postal_code,city,street1,resource_path,created_at,updated_at) VALUES
    ('01H7C7AGSH4KJPAQYXP0Q0FYPB','01GX5MP57FS52F1MGDQ7SJ6CDS','01GWTEE79PQY8C0JTWC1G08SX5','dwh-postal-test','dwh-city-test','dwh-street1-test','-2147483642',now(), now()) ON CONFLICT DO NOTHING;
