\connect invoicemgmt;

INSERT INTO public.payment (invoice_id,payment_id,student_id,payment_status,payment_method,payment_due_date,payment_expiry_date,payment_date,amount,result_code,resource_path,created_at,updated_at,deleted_at) VALUES
    ('01GWHQV538V3AP0WVDFMZ0TJ2W','01GWSN9YWVPWE3JDW3EJSDN8T7','01GWHQY4BBRYZ7G0XDEQN22TN0','PAYMENT_PENDING','CONVENIENCE_STORE','2023-04-01 00:15:00+07','2023-04-02 00:15:00+07','2023-04-03 00:15:00+07',500.00,null,'-2147483642',now(),now(),NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.invoice_bill_item(
	invoice_bill_item_id, invoice_id, bill_item_sequence_number,past_billing_status, resource_path, created_at,  updated_at)
	VALUES ('01GWTE5E2HJKYR2Y9H7NJ9FM86', '01GWHQV538V3AP0WVDFMZ0TJ2W', 1,'BILLING_STATUS_BILLED', '-2147483642', now(), now()) ON CONFLICT DO NOTHING;


INSERT INTO public.student_payment_detail(
	student_payment_detail_id, student_id, payer_name, payer_phone_number, payment_method, created_at, updated_at, resource_path)
	VALUES ('01GWTEE79PQY8C0JTWC1G08SX5', '01GWHQY4BBRYZ7G0XDEQN22TN0', 'test-payer', '0123416', 'CONVENIENCE_STORE', now(), now(), '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.users (user_id, email, country, name, user_group, updated_at, created_at, resource_path)
VALUES ('01GX5MP57FS52F1MGDQ7SJ6CDS', 'staff@manabie.com', 'COUNTRY_VN', 'Staff name', 'USER_GROUP_TEACHER', '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_basic_info (user_id, updated_at, created_at, resource_path)
VALUES ('01GX5MP57FS52F1MGDQ7SJ6CDS',  '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642')
ON CONFLICT DO NOTHING;


INSERT INTO public.student_payment_detail_action_log(
	student_payment_detail_action_id, student_payment_detail_id, user_id, action, action_detail, resource_path, created_at, updated_at)
	VALUES ('01GWTEK28312EJHK2AGTGAJRD1', '01GWTEE79PQY8C0JTWC1G08SX5', '01GX5MP57FS52F1MGDQ7SJ6CDS', 'UPDATED_BILLING_DETAILS', '{"new":{"bank_account":null,"billing_address":{"city":"01GWS8XHDB9TPCW7Y32AGYQ9MF-city-updated_at_1680180436364","street1":"01GWS8XHDB9TPCW7Y32AGYQ9MF-street_1-updated_at_1680180436364","street2":"","payer_name":"01GWS8XHDB9TPCW7Y32AGYQ9MF-payer_name-updated_at_1680180436364","postal_code":"01GWS8XHDB9TPCW7Y32AGYQ9MF-postal_code-updated_at_1680180436364","prefecture_code":"prefecture-code-01GWS8XQC9VQCSHRAGAWRE2K4V","payer_phone_number":"01GWS8XHDB9TPCW7Y32AGYQ9MF-payer_phone_number-updated_at_1680180436364"}},"previous":{"bank_account":null,"billing_address":{"city":"01GWS8XHDB9TPCW7Y32AGYQ9MF-city","street1":"01GWS8XHDB9TPCW7Y32AGYQ9MF-street_1","street2":"01GWS8XHDB9TPCW7Y32AGYQ9MF-street_2","payer_name":"01GWS8XHDB9TPCW7Y32AGYQ9MF-payer_name","postal_code":"01GWS8XHDB9TPCW7Y32AGYQ9MF-postal_code","prefecture_code":"prefecture-code-01GWS8XHM6DZ1BZHVZK8J64JQT","payer_phone_number":"01GWS8XHDB9TPCW7Y32AGYQ9MF-payer_phone_number"}}}', '-2147483642', now(), now()) ON CONFLICT DO NOTHING;
