INSERT INTO permission_role (permission_id,role_id,created_at,updated_at,resource_path) VALUES
    ('01GZZJHVM0DXT9CVXG08ER0051','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623'), -- bank_account.read
    ('01GZZJJ0R1A97SZ92WTY5RE6RS','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623'), -- bank_account.write

    ('01GZZJJ5FRGK5WRATE0KZZSHY1','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623'), -- billing_address.read
    ('01GZZJJBEEMQDJBZX9753ZQZ79','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623'), -- billing_address.write

    ('01GZZJKT5F2C02KE69P3Y80W57','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623'), -- student_payment_detail.read
    ('01GZZJKZ93VMKGJ97AFPV7SN03','01GZZJQD0Q06AP9X3HC3ZHMQ00','now()','now()','-2147483623') ON CONFLICT DO NOTHING; -- student_payment_detail.write