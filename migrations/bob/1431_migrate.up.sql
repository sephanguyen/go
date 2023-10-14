-- 1. ------(-2147483629)-----Withus (Managara High School)----------
INSERT INTO public.permission (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68J7T2PCR', 'master.course.read', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68P78NVZ6', 'master.course.write', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;

-- permission role for master.location.read
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC0', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC1', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC2', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC3', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC4', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC5', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC6', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC7', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC8', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFMNZZS2HKS7Y1J6EQPGEBC9', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GJZNHK0NJSK9Z6ASGSPXRX2R', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68J7T2PCR', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;

-- permission role master.location.write
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68P78NVZ6', '01GFMNZZS2HKS7Y1J6EQPGEBC2', now(), now(), '-2147483629'),
('01GM2DS7WP4S4JY7Q68P78NVZ6', '01GFMNZZS2HKS7Y1J6EQPGEBC5', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;

-- delete redundant notification permission
DELETE FROM granted_permission 
WHERE permission_id = '01GFWPT9WFAZBSR84N14HRTGN1' OR permission_id = '01GFWPT9WFAZBSR84N14HRTGN2' AND resource_path = '-2147483629';
DELETE FROM permission_role WHERE permission_id = '01GFWPT9WFAZBSR84N14HRTGN1'OR permission_id = '01GFWPT9WFAZBSR84N14HRTGN2'AND resource_path = '-2147483629';
DELETE FROM "permission" WHERE permission_id = '01GFWPT9WFAZBSR84N14HRTGN1' OR permission_id = '01GFWPT9WFAZBSR84N14HRTGN2' AND resource_path = '-2147483629';

-- add missing notification permission for notification cronjob role
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GGVXWS0VDR3RDAM114X2DBVF', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629'),
    ('01GGVXWS0Z5NBY2VKWBCXX48G6', '01GFWQC67NF55GYEXBNRR12AD1', now(), now(), '-2147483629')
ON CONFLICT DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S1')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;


-- 2. ------(-2147483630)-----Withus (Managara Base)----------
INSERT INTO public.permission (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68EXGMKAM', 'master.course.read', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68H5QYBZN', 'master.course.write', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;

-- permission role master.location.read
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F0', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F1', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F2', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F3', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F4', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F5', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F6', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F7', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F8', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFMMFRXDZHGVC3YWC7J668F9', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GJZNHK0NJSK9Z6ASGWXHXS7X', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68EXGMKAM', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;

-- permission role master.location.write
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM2DS7WP4S4JY7Q68H5QYBZN', '01GFMMFRXDZHGVC3YWC7J668F2', now(), now(), '-2147483630'),
('01GM2DS7WP4S4JY7Q68H5QYBZN', '01GFMMFRXDZHGVC3YWC7J668F5', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;

-- delete redundant notification permission
DELETE FROM granted_permission WHERE permission_id = '01GFWPV54P3P49ZPT3BB069X21' OR permission_id = '01GFWPV54P3P49ZPT3BB069X22' AND resource_path = '-2147483630';
DELETE FROM permission_role WHERE permission_id = '01GFWPV54P3P49ZPT3BB069X21' OR permission_id = '01GFWPV54P3P49ZPT3BB069X22' AND resource_path = '-2147483630';
DELETE FROM "permission" WHERE permission_id = '01GFWPV54P3P49ZPT3BB069X21' OR permission_id = '01GFWPV54P3P49ZPT3BB069X22' AND resource_path = '-2147483630';

-- add missing notification permission for notification cronjob role
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GGVRH1MRCTHP8G0S9AWF3JFZ', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630'),
    ('01GGVRH1MWF3GD3BQ1QD2TYJEX', '01GFWQC67NF55GYEXBNRR12AD2', now(), now(), '-2147483630')
ON CONFLICT DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S2')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;


-- 3. ------(-2147483631)-----Eishinkan----------
INSERT INTO public.permission (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
('01GM10XAQ7GGXTT6KFCAEXZFH5', 'master.course.read', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCBKN7615', 'master.course.write', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;

-- permission role master.location.read
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N670', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N671', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N672', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N673', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N674', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N675', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N676', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N677', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N678', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GDWSMJS45TK897ZA6TN2N679', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GJZNHK0NJSK9Z6ASGYW5J1XY', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCAEXZFH5', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;

-- permission role master.location.write
INSERT INTO public.permission_role (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
('01GM10XAQ7GGXTT6KFCBKN7615', '01GDWSMJS45TK897ZA6TN2N672', now(), now(), '-2147483631'),
('01GM10XAQ7GGXTT6KFCBKN7615', '01GDWSMJS45TK897ZA6TN2N675', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;

-- delete redundant notification permission
DELETE FROM granted_permission WHERE permission_id = '01GFWPW8R69G0MXSTWSPACHJ31' OR permission_id = '01GFWPW8R69G0MXSTWSPACHJ32' AND resource_path = '-2147483631';
DELETE FROM permission_role WHERE permission_id = '01GFWPW8R69G0MXSTWSPACHJ31' OR permission_id = '01GFWPW8R69G0MXSTWSPACHJ32' AND resource_path = '-2147483631';
DELETE FROM "permission" WHERE permission_id = '01GFWPW8R69G0MXSTWSPACHJ31' OR permission_id = '01GFWPW8R69G0MXSTWSPACHJ32' AND resource_path = '-2147483631';

-- add missing notification permission for notification cronjob role
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GGVJB2C4F9FR13NTDNZC9KMD', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631'),
    ('01GGVJB2C5R3Y7BW7VXC689S04', '01GFWQC67NF55GYEXBNRR12AD3', now(), now(), '-2147483631')
ON CONFLICT DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GFWR7AGF3KQ6SHDTMDH7S8S3')
ON CONFLICT ON CONSTRAINT granted_permission__pk
DO UPDATE SET user_group_name = excluded.user_group_name;
