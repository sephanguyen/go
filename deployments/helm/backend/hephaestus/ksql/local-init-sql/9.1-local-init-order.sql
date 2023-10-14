\connect fatima;

INSERT INTO public.product
(product_id, "name", "product_type", tax_id, available_from, available_until, remarks, custom_billing_period, billing_schedule_id, disable_pro_rating_flag, is_archived, updated_at, created_at, resource_path, is_unique, product_tag, product_partner_id)
VALUES('01GPA8QPA0XZE30DFQ3389CVZN', 'KEC DWH TEST2', 'PRODUCT_TYPE_FEE', NULL, '2022-06-23 00:00:00.000', '2025-06-22 23:59:59.999', 'Product generated using scripts on 6/23/2023', '2023-07-23 00:00:00.000', NULL, false, false, '2023-06-23 11:27:06.641', '2023-06-23 11:27:06.641', '-2147483642', false, NULL, NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.discount (discount_id,name,discount_type,discount_amount_type,discount_amount_value,available_from,available_until,resource_path,created_at,updated_at) VALUES
    ('01GP2Z2QJHFRGC47GTHFVYTQEM','test-discount-name','test-discount-type','test-amount-type',500.00,now(),now(),'-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.product_discount (discount_id,product_id,resource_path,created_at) VALUES
    ('01GP2Z2QJHFRGC47GTHFVYTQEM','01GPA8QPA0XZE30DFQ3389CVZN','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.upcoming_student_course (upcoming_student_package_id,student_id,course_id,location_id,student_package_id,package_type,course_slot,course_slot_per_week,weight,student_start_date,student_end_date,resource_path,created_at,updated_at) VALUES
    ('01GPA9H3GTNQKEP0WE0R7FSJVF','01GS424P90TKQD1JHW6TJH1DTQ','01GPAMKZA30TR2XTGMRFA939C4','01GPAFES0TE7DKMP0KH7SSVJV7','ce07d58f-e274-4a64-ac2f-a4907289a40d','PACKAGE_TYPE_ONE_TIME',null,null,2,'2021-12-31T23:00:00+00:00','2025-08-17T23:00:00+00:00','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.student_course (student_id,course_id,location_id,student_package_id,package_type,course_slot,course_slot_per_week,weight,student_start_date,student_end_date,resource_path,created_at,updated_at) VALUES
    ('01H5M8SNCEKFHDR498RENSJ9MW','01GWK5JB12FJC16STAMWWBCOURSE1','01GV032YZ8FA4JGEAR4XXQX33','ce07d58f-e274-4a64-ac2f-a4907289a40d','PACKAGE_TYPE_ONE_TIME',null,null,2,'2021-12-31T23:00:00+00:00','2025-08-17T23:00:00+00:00','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.course_access_paths (course_id,location_id,resource_path,created_at,updated_at) VALUES
    ('01GWK5JB12FJC16STAMWWBCOURSE1','01GV032YZ8FA4JGEAR4XXQX33','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.student_product
(student_product_id, student_id, product_id, upcoming_billing_date, start_date, end_date, product_status, approval_status, updated_at, created_at, deleted_at, resource_path, location_id, updated_from_student_product_id, updated_to_student_product_id, student_product_label, is_unique, root_student_product_id, is_associated, version_number)
VALUES('202307TDN3F2PSCGJVYGP0078', '01H5M8SNCEKFHDR498RENSJ9MW', '01GPA8QPA0XZE30DFQ3389CVZN', NULL, '2022-11-04 15:59:16.000', '2022-11-03 16:08:24.000', 'ORDERED', NULL, '2022-11-04 16:09:31.957', '2022-11-04 15:59:30.677', NULL, '-2147483642', '01GV032YZ8FA4JGEAR4XXQX33', NULL, NULL, 'UPDATE_SCHEDULED', false, NULL, false, 0)  ON CONFLICT DO NOTHING;

INSERT INTO public.student_associated_product (student_product_id,associated_product_id,resource_path,created_at,updated_at) VALUES
    ('202307TDN3F2PSCGJVYGP004','202307TDN3F2PSCGJVYGP0078','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.file
(file_id, file_name, file_type, download_link, created_at, updated_at, deleted_at, resource_path)
VALUES ('01GM7PGBTEXCGFPNVS8C54Q8BH','ENROLLMENT','PDF', 'https://manabie.com', '2022-12-14T06:52:44.404654+00:00','2022-12-14T06:52:44.404654+00:00', null,'-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.order_item
(order_id, product_id, order_item_id, discount_id, start_date, created_at, student_product_id, product_name, effective_date, cancellation_date, end_date, updated_at, deleted_at, resource_path)
VALUES ('01H5KY9ZGCNCBW4X1SB6X89MX3','01GPA8QPA0XZE30DFQ3389CVZN','01GM7PGBTEXCGFPNVS8C54Q9TX','01GP2Z2QJHFRGC47GTHFVYTQEM','2022-12-14T06:52:44.404654+00:00','2022-12-14T06:52:44.404654+00:00','202307TDN3F2PSCGJVYGP0078','product name fatima','2023-7-14T06:52:44.404654+00:00', null, '2024-12-14T06:52:44.404654+00:00', null, null, '-2147483642') ON CONFLICT DO NOTHING;
