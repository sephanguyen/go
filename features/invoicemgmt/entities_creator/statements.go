package entitiescreator

const (
	InsertLocationStmt = `INSERT INTO locations (
		location_id,
		access_path,
		name,
		location_type,
		partner_internal_id,
		updated_at,
		created_at
	) VALUES ($1, $2, $3, $4, $5, now(), now())
	`

	InsertInvoiceBillItemStmt = `INSERT INTO invoice_bill_item (
		invoice_id,
		bill_item_sequence_number,
		past_billing_status,
		created_at
	) VALUES ($1, $2, $3, now()) 
	RETURNING invoice_bill_item_id
	`

	InsertBillingScheduleStmt = `INSERT INTO billing_schedule (
		billing_schedule_id,
		name,
		remarks,
		is_archived,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, now(), now())
	RETURNING billing_schedule_id
	`

	InsertBillingSchedulePeriodStmt = `INSERT INTO billing_schedule_period (
		billing_schedule_period_id,
		name,
		billing_schedule_id,
		start_date,
		end_date,
		billing_date,
		is_archived,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
	RETURNING billing_schedule_period_id
	`

	InsertProductStmt = `INSERT INTO product (
		product_id,
		name,
		product_type,
		tax_id,
		available_from,
		available_until,
		remarks,
		custom_billing_period,
		billing_schedule_id,
		disable_pro_rating_flag,
		is_archived,
		updated_at,
		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
	RETURNING product_id
	`

	InsertTaxStmt = `INSERT INTO tax(
		tax_id,
		name,
		tax_percentage,
		tax_category,
		default_flag,
		is_archived,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	RETURNING tax_id
	`

	InsertStudentProductStmt = `INSERT INTO student_product (
		student_product_id,
		student_id,
		product_id ,
		upcoming_billing_date,
		start_date,
		end_date,
		product_status,
		approval_status,
		updated_at,
		created_at,
		deleted_at,
		location_id
	) VALUES ($1, $2, $3, now(), now(), now(), $4, $5, now(), now(), NULL, $6)
	`

	InsertBankStmt = `INSERT INTO bank (
		bank_id,
		bank_code,
		bank_name ,
		bank_name_phonetic,
		is_archived
	) VALUES ($1, $2, $3, $4, $5)
	`

	InsertGrantedRoleAccessPathStmt = `INSERT INTO granted_role_access_path (
		granted_role_id,
		location_id,
		created_at,
		updated_at
	) VALUES ($1, $2, now(), now())
	`

	InsertPrefectureStmt = `INSERT INTO prefecture (
		prefecture_id,
		prefecture_code,
		country,
		name,
		updated_at,
		created_at
	) VALUES ($1, $2, $3, $4, now(), now())
	`

	InsertOrderStmt = `INSERT INTO "order" (
		order_id,
		student_id,
		order_comment,
		order_status,
		student_full_name,
		order_type,
		location_id,
		is_reviewed,
		created_at,
		updated_at
	)  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
	`

	InsertUserBasicInfoStmt = `INSERT INTO user_basic_info (
		user_id,
		name,
		first_name,
		last_name,
		full_name_phonetic,
		first_name_phonetic,
		last_name_phonetic,
		current_grade,
		grade_id,
		updated_at,
		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
	`
)
