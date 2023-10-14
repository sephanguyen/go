\connect fatima;

INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV032YZ8FA4JGEAR4XXQX33', 'Location DWH test', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642', NULL, '', '', NULL, false, '01GV032YZ8FA4JGEAR4XXQX6L3');

INSERT INTO public.courses
(course_id, "name", updated_at, created_at, deleted_at, resource_path)
VALUES('01GWK5JB12FJC16STAMWWBCOURSE1', 'course DWH test', '2023-07-19 17:49:25.803', '2023-07-19 17:49:25.803',  NULL, '-2147483642');


INSERT INTO public.tax (tax_id,name,tax_percentage,tax_category,resource_path,created_at,updated_at) VALUES
    ('01H5KYK8J8X56E8W7MFWQ4FE41','test-tax-name',5,'TAX_CATEGORY_NONE','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.students(student_id, current_grade, updated_at, created_at, resource_path) VALUES
    ('01H5M8SNCEKFHDR498RENSJ9MW', 1,  now(), now(), '-2147483642');

INSERT INTO public.order (order_id,student_id,location_id,order_status,student_full_name,resource_path,created_at,updated_at) VALUES
    ('01H5KY9ZGCNCBW4X1SB6X89MX3','01H5M8SNCEKFHDR498RENSJ9MW','01GV032YZ8FA4JGEAR4XXQX33','test-order-status','test-student','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.bill_item (order_id,product_description,bill_type,billing_status,final_price,student_id,location_id,resource_path,created_at,updated_at,tax_id) VALUES
    ('01H5KY9ZGCNCBW4X1SB6X89MX3','test-product-descript','test-bill-type','test-bill-status',500.00,'01H5M8SNCEKFHDR498RENSJ9MW','01GV032YZ8FA4JGEAR4XXQX33','-2147483642',now(),now(),'01H5KYK8J8X56E8W7MFWQ4FE41') ON CONFLICT DO NOTHING;

INSERT INTO public.discount (discount_id,name,discount_type,discount_amount_type,discount_amount_value,available_from,available_until,resource_path,created_at,updated_at) VALUES
    ('01H5SNTBX4JZJH2RRCKWHNCYFP','test-discount-name','test-discount-type','test-amount-type',500.00,now(),now(),'-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.order_item (order_item_id,order_id,discount_id,resource_path,created_at) VALUES
    ('01H5SP0YQ3T4WWTPK99MQ2TD26','01H5KY9ZGCNCBW4X1SB6X89MX3','01H5SNTBX4JZJH2RRCKWHNCYFP','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.product (product_id,name,product_type,available_from,available_until,resource_path,created_at,updated_at) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','test-name','test-type',now(),now(),'-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.material (material_id,material_type,resource_path) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','test-type','-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.billing_schedule (billing_schedule_id,name,resource_path,created_at,updated_at) VALUES
    ('01H5SPXY7B0HHEWMVYEHY8FXN4','test-name','-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.billing_schedule_period (billing_schedule_period_id,name,billing_schedule_id,start_date,end_date,billing_date,resource_path,created_at,updated_at) VALUES
    ('01H5SQ7KGE0BR6QBMH966F71SW','test-name','01H5SPXY7B0HHEWMVYEHY8FXN4',now(),now(),now(),'-2147483642',now(),now()) ON CONFLICT DO NOTHING;

INSERT INTO public.billing_ratio (billing_ratio_id,start_date,end_date,billing_schedule_period_id,billing_ratio_numerator,billing_ratio_denominator,resource_path,created_at,updated_at) VALUES
    ('01H5SQA4J8FXNWYAMTXKM3GTGT',now(),now(),'01H5SQ7KGE0BR6QBMH966F71SW',1,2,'-2147483642',now(),now()) ON CONFLICT DO NOTHING;



INSERT INTO public.leaving_reason
(leaving_reason_id, "name", leaving_reason_type, remark, is_archived, updated_at, created_at, resource_path)
VALUES('202307TDN3F2PSCGJVYGP001', 'KEC DWH TEST', 'LEAVING_REASON_TYPE_LOA', NULL, false, '2023-04-12 12:45:49.133',
 '2023-04-12 12:45:49.133', '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.product
(product_id, "name", "product_type", tax_id, available_from, available_until, remarks, custom_billing_period, billing_schedule_id, disable_pro_rating_flag, is_archived, updated_at, created_at, resource_path, is_unique, product_tag, product_partner_id)
VALUES('202307TDN3F2PSCGJVYGP006', 'KEC DWH TEST', 'PRODUCT_TYPE_FEE', NULL, '2022-06-23 00:00:00.000', '2025-06-22 23:59:59.999', 'Product generated using scripts on 6/23/2023', '2023-07-23 00:00:00.000', NULL, false, false, '2023-06-23 11:27:06.641', '2023-06-23 11:27:06.641', '-2147483642', false, NULL, NULL)  ON CONFLICT DO NOTHING;


INSERT INTO public.student_product
(student_product_id, student_id, product_id, upcoming_billing_date, start_date, end_date, product_status, approval_status, updated_at, created_at, deleted_at, resource_path, location_id, updated_from_student_product_id, updated_to_student_product_id, student_product_label, is_unique, root_student_product_id, is_associated, version_number)
VALUES('202307TDN3F2PSCGJVYGP004', '01H5M8SNCEKFHDR498RENSJ9MW', '202307TDN3F2PSCGJVYGP006', NULL, '2022-11-04 15:59:16.000', '2022-11-03 16:08:24.000', 'ORDERED', NULL, '2022-11-04 16:09:31.957', '2022-11-04 15:59:30.677', NULL, '-2147483642', '01GV032YZ8FA4JGEAR4XXQX33', NULL, NULL, 'UPDATE_SCHEDULED', false, NULL, false, 0)  ON CONFLICT DO NOTHING;


INSERT INTO public.bill_item_course
(bill_item_sequence_number, course_id, course_name, course_weight, course_slot, course_slot_per_week, created_at, resource_path)
VALUES(1, '01GWK5JB12FJC16STAMWWBCOURSE1', 'KEC DWH TEST', 1, NULL, NULL, '2022-11-04 15:59:31.605', '-2147483642')  ON CONFLICT DO NOTHING;


INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id,
platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, 
last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic)
VALUES('01GWK5JB12FJC16STAMWWBUSER1', 'COUNTRY_JP', 'vctqs1 student', NULL, NULL, 'thu.vo+e2estudent@manabie.com', NULL, false, 'USER_GROUP_STUDENT', '2023-03-09 22:24:18.833', '2023-03-09 22:24:18.833', false, NULL, NULL, false, false, NULL, NULL, '-2147483642', NULL, NULL, NULL, 'student', 'vctqs1', '', '', '') ON CONFLICT DO NOTHING;


INSERT INTO public.order_item_course
(order_id, package_id, course_id, course_name, course_slot, course_slot_per_week, created_at, updated_at, resource_path, order_item_course_id)
VALUES('01H5KY9ZGCNCBW4X1SB6X89MX3', '01GH0X07H8CWDMPNG166Z9H7ZD', '01GWK5JB12FJC16STAMWWBCOURSE1', 'Peanuts', NULL, NULL, '2022-11-04 15:59:31.027', '2022-11-04 15:59:31.027', '-2147483642', '01H5KY9ZGCNCBW4X1SB6X89MX3')  ON CONFLICT DO NOTHING;

INSERT INTO public.package_quantity_type_mapping
(package_type, quantity_type, created_at, resource_path)
VALUES('PACKAGE_TYPE_ONE_TIME', 'QUANTITY_TYPE_COURSE_WEIGHT', '2022-12-20 11:31:57.257', '-2147483642')  ON CONFLICT DO NOTHING;

INSERT INTO public.order_action_log
(order_action_log_id, user_id, order_id, "action", "comment", created_at, updated_at, resource_path)
VALUES(6360, '01GWK5JB12FJC16STAMWWBUSER1', '01H5KY9ZGCNCBW4X1SB6X89MX3', 'ORDER_ACTION_SUBMITTED', '', '2022-11-04 15:59:30.665', '2022-11-04 15:59:30.665', '-2147483642')  ON CONFLICT DO NOTHING;

INSERT INTO public.accounting_category
(accounting_category_id, "name", remarks, is_archived, updated_at, created_at, resource_path)
VALUES('202307TDN3F2PSCGJVYGP007', 'KEC DWH TEST', NULL, false, '2023-01-30 13:37:27.921', '2023-01-30 13:37:27.921', '-2147483642')  ON CONFLICT DO NOTHING;



INSERT INTO public.product_setting
(product_id, is_enrollment_required, created_at, updated_at, resource_path, is_pausable, is_added_to_enrollment_by_default, is_operation_fee)
VALUES('202307TDN3F2PSCGJVYGP006', true, '2022-09-06 16:38:54.908', '2022-09-06 16:38:54.908', '-2147483642', true, false, false)  ON CONFLICT DO NOTHING;

INSERT INTO public.package
(package_id, package_type, max_slot, package_start_date, package_end_date, resource_path)
VALUES('202307TDN3F2PSCGJVYGP006', 'type', 1, '2023-07-21 13:59:58.580', '2023-07-21 13:59:58.580', '-2147483642')  ON CONFLICT DO NOTHING;


INSERT INTO public.package_course
(package_id, course_id, mandatory_flag, course_weight, created_at, resource_path, max_slots_per_course)
VALUES('202307TDN3F2PSCGJVYGP006', '01GWK5JB12FJC16STAMWWBCOURSE1', false, 3, '2022-10-31 12:58:53.727', '-2147483642', 1)  ON CONFLICT DO NOTHING;



INSERT INTO public.product_location
(product_id, location_id, created_at, resource_path)
VALUES('202307TDN3F2PSCGJVYGP006', '01GV032YZ8FA4JGEAR4XXQX33', '2022-08-03 15:07:15.629', '-2147483642')  ON CONFLICT DO NOTHING;

INSERT INTO public.product(
	product_id, name, product_type, tax_id, available_from, available_until, disable_pro_rating_flag, is_archived, is_unique, created_at, updated_at, resource_path)
	VALUES ('01G9VAT3372YZ9P03WV82KJ3G3','fee one time','FEE_TYPE_ONE_TIME', null,'2022-08-12 10:31:18.725', '2025-08-12 10:31:18.725', false, false, false, now(), now(), '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.product(
	product_id, name, product_type, tax_id, available_from, available_until, disable_pro_rating_flag, is_archived, is_unique, created_at, updated_at, resource_path)
	VALUES ('01G9VR7D5Y9CA58P5H85QCJSHG','package one time','PACKAGE_TYPE_ONE_TIME', null,'2022-08-12 10:31:18.725', '2025-08-12 10:31:18.725', false, false, false, now(), now(), '-2147483642') ON CONFLICT DO NOTHING;


INSERT INTO public.fee(
	fee_id, fee_type, resource_path)
	VALUES ('01G9VAT3372YZ9P03WV82KJ3G3', 'FEE_TYPE_ONE_TIME', '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.package(
	package_id, package_type, max_slot, package_start_date, package_end_date, resource_path)
	VALUES ('01G9VR7D5Y9CA58P5H85QCJSHG', 'PACKAGE_TYPE_ONE_TIME', 1, '2022-08-12 10:31:18.725', '2025-08-12 10:31:18.725','-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.product_grade (product_id,grade_id,resource_path,created_at) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','01H5TZP5Q977TYQRSS5AV9PNAB','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.product_price (product_price_id,product_id,price,resource_path,created_at) VALUES
    (99999,'01H5SPMM2FJKYK6ADE2MPQCM6V',50.00,'-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.fee(
	fee_id, fee_type, resource_path)
	VALUES ('01H5SPMM2FJKYK6ADE2MPQCM6V', 'FEE_TYPE_ONE_TIME', '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.package(
	package_id, package_type, max_slot, package_start_date, package_end_date, resource_path)
	VALUES ('01H5SPMM2FJKYK6ADE2MPQCM6V', 'PACKAGE_TYPE_ONE_TIME', 1, '2022-08-12 10:31:18.725', '2025-08-12 10:31:18.725', '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.courses(
	course_id, name, created_at,updated_at,resource_path)
	VALUES ('01H5V0RVPN70TZ8HQ2P5PCG09K','test-course',now(),now(), '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.package_course_fee (package_id,course_id,fee_id,resource_path,created_at) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','01H5V0RVPN70TZ8HQ2P5PCG09K', '01H5SPMM2FJKYK6ADE2MPQCM6V','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.package_course_material (package_id,course_id,material_id,resource_path,created_at) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','01H5V0RVPN70TZ8HQ2P5PCG09K', '01H5SPMM2FJKYK6ADE2MPQCM6V','-2147483642',now()) ON CONFLICT DO NOTHING;

INSERT INTO public.accounting_category (accounting_category_id,name,resource_path,created_at, updated_at) VALUES
    ('01H5V31TZ2G00WKZEBPQQZZN4S','test-accounting-category-name','-2147483642',now(), now()) ON CONFLICT DO NOTHING;

INSERT INTO public.product_accounting_category (product_id,accounting_category_id,resource_path,created_at) VALUES
    ('01H5SPMM2FJKYK6ADE2MPQCM6V','01H5V31TZ2G00WKZEBPQQZZN4S','-2147483642',now()) ON CONFLICT DO NOTHING;
