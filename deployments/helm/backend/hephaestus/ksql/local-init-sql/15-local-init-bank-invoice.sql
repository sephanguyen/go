\connect invoicemgmt;

INSERT INTO bank(bank_id, bank_code, bank_name, bank_name_phonetic, is_archived, created_at, updated_at, resource_path)
    VALUES ('01H7BZ2904XXJV9QB7GS8HP81R', '1234', 'init bank name', 'init bank name phonetic', false, now(), now(), '-2147483642') ON CONFLICT DO NOTHING;


INSERT INTO bank_branch(bank_branch_id, bank_id, bank_branch_code, bank_branch_name, bank_branch_phonetic_name, is_archived, created_at, updated_at, resource_path)
    VALUES ('01H7BZ5Z1CKPHM6BJKDY7R0W7C', '01H7BZ2904XXJV9QB7GS8HP81R', '123', 'init bank branch name', 'init bank branch name phonetic', false, now(), now(), '-2147483642')  ON CONFLICT DO NOTHING;

INSERT INTO public.partner_bank(partner_bank_id, bank_number, bank_name, bank_branch_number, bank_branch_name, 
    deposit_items, account_number, remarks, is_archived, created_at, updated_at, deleted_at, resource_path, consignor_code, consignor_name, is_default, record_limit)
	VALUES ('01H7CKGWYZYSF365T5AAG5794K', '1234', 'init partner bank name', '123', 'init partner bank branch name', 'init deposit items', '1234567', 
        'init remarks', false, now(), noW(), null, '-2147483642', 'init consignor code', 'init consignor name', false, 0)  ON CONFLICT DO NOTHING;

INSERT INTO public.bank_mapping(bank_mapping_id, bank_id, partner_bank_id, remarks, is_archived, created_at, updated_at, deleted_at, resource_path)
	VALUES ('01H7CKNE32Z02CWTV9SQ0W68GN', '01H7BZ2904XXJV9QB7GS8HP81R', '01H7CKGWYZYSF365T5AAG5794K', 'init remarks', false, now(), noW(), null, '-2147483642')  ON CONFLICT DO NOTHING;

INSERT INTO bank_account(bank_account_id, student_payment_detail_id, student_id, bank_branch_id, bank_account_number, bank_account_holder, bank_account_type, bank_id, resource_path, created_at, updated_at)
    VALUES ('01H7DD3RP2CHYW4N38BG7TZQBN', '01GWTEE79PQY8C0JTWC1G08SX5', '01GWHQY4BBRYZ7G0XDEQN22TN0', '01H7BZ5Z1CKPHM6BJKDY7R0W7C','1234567', 'dwh-test-account-holder','test-bank-account-type-dwh', '01H7BZ2904XXJV9QB7GS8HP81R', '-2147483642', now(),now())  ON CONFLICT DO NOTHING;
