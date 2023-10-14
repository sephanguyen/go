CREATE TABLE IF NOT EXISTS public.organizations (
	organization_id text NOT NULL,
	tenant_id text NULL,
	"name" text NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	domain_name text NULL,
	logo_url text NULL,
	country text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	scrypt_signer_key text NULL,
	scrypt_salt_separator text NULL,
	scrypt_rounds text NULL,
	scrypt_memory_cost text NULL,
	CONSTRAINT organization__domain_name__check CHECK ((domain_name ~ '^[^-\s].[a-z0-9-]*$'::text)),
	CONSTRAINT organization__domain_name__un UNIQUE (domain_name),
	CONSTRAINT organizations__pk PRIMARY KEY (organization_id),
	CONSTRAINT organizations__tenant_id__un UNIQUE (tenant_id)
);

INSERT INTO organizations (organization_id,tenant_id,"name",resource_path,domain_name,logo_url,country,created_at,updated_at,deleted_at,scrypt_signer_key,scrypt_salt_separator,scrypt_rounds,scrypt_memory_cost) VALUES
	 ('-2147483629','withus-managara-hs-2391o','Managara High School','-2147483629','managara-hs','https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant-managara-hs-logo.png','COUNTRY_JP','2022-10-18 13:54:45.998875+07','2022-10-18 13:54:45.998875+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483630','withus-managara-base-mrkvu','Managara Base','-2147483630','managara-base','https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant-managara-base-logo.png','COUNTRY_JP','2022-10-18 13:42:26.378883+07','2022-10-18 13:42:26.378883+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483631','eishinkan-de5sd','Eishinkan','-2147483631','eishinkan-group','https://storage.googleapis.com/prod-tokyo-backend/user-upload/eishinkan.png','COUNTRY_JP','2022-09-27 09:57:33.726528+07','2022-09-27 09:57:33.726528+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483632','manabie-kael-6ys95','Manabie Data Leak check','-2147483632','manabie-kael','','COUNTRY_JP','2022-08-15 16:34:16.627812+07','2022-08-15 16:34:16.627812+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483633','manabie-tech-xbglh','Manabie Tech','-2147483633','manabie-tech','','COUNTRY_JP','2022-08-15 16:25:37.885287+07','2022-08-15 16:25:37.885287+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483634','manabie-b2c-testing-mg761','Manabie B2C Internal','-2147483634','manabie-b2c-internal','','COUNTRY_VN','2022-07-25 18:37:54.391983+07','2022-07-25 18:37:54.391983+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483635','prod-kec-demo-fkl3m','KEC Demo','-2147483635','kec-demo','','COUNTRY_JP','2022-07-25 11:24:01.405225+07','2022-07-25 11:24:01.405225+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483637','prod-manabie-demo-v2amk','Manabie Demo','-2147483637','manabie-demo','https://storage.googleapis.com/prod-tokyo-backend/user-upload/manabie_ic_splash.png','COUNTRY_JP','2022-06-24 18:11:23.046873+07','2022-06-24 18:11:23.046873+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483638','prod-e2e-hcm-8q86x','E2E HCM','-2147483638','e2e-hcm','https://storage.googleapis.com/prod-tokyo-backend/user-upload/manabie_ic_splash.png','COUNTRY_JP','2022-06-24 18:11:22.153873+07','2022-06-24 18:11:22.153873+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483639','prod-e2e-tokyo-2k4xb','E2E Tokyo','-2147483639','e2e-tokyo',NULL,'COUNTRY_JP','2022-05-26 09:59:59.627619+07','2022-05-26 09:59:59.627619+07',NULL,'R/OeaTgyC8/FGsZg2drrExsCnZarErCiGm9YBLDnKROqnSCRh7yc5z2r4NguoldXXmstmSyyWyW7
RBvfxY8MIREariB1eCjrxVF/o7fJ+/VA4mRRjekx8CsIMofEFMfk','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483640','prod-nsg-flbh7','NSG School','-2147483640','nsg',NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'Q+AVOjqk9Uih+FA2MZqufJHCqkHPBAveYoYBjydzlbXzx2ph/8NTQDIR9afRIJDZO6xI6MnfUimA
z40oTeasGeBhwZmjloj5sNHF1wKKljus1YfdJ4GctoxMp0vjF5R1','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483641','prod-aic-u3d1m','AIC School','-2147483641','aic-oshu','https://storage.googleapis.com/prod-tokyo-backend/user-upload/aic_logo.png',NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'VYV4NRD+Qa5EXQHeARXlb7FwK7kC5rLNoMlWB5nzQk3CkicMvfUU2HXWrm0zqHoxH41u8E0R6bGk
6+pUefgYw63CA1L2z60o68OcTWuntSv8If51xAaCeeqRt+xWJJPi','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483642','prod-kec-58ww0','KEC School','-2147483642','kec',NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'QoCdNpAWFRENODNG0OR19J3762sq1n44YoiS5aNqzlyU2/XkVJ6sgrpyJ9y5kLoa6zGj/0X6HUpg
yH6t0lUtsUIs7hUUo6mPb1BnADz/sGhF024ao/VoqVN3EBuh6wWE','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483643','prod-ga-uq2rq','Bestco','-2147483643','bestco',NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'zN+zSp35y9DvviPOOoSgFwQm6Svb8vqzFtUrhmAxDtfkf/Ve4Rh0iGMnjCfhc5ukhM9Ct8VBvo0J
Qf5NIbHeBBM9R28uTIQKKYXivQKS6I/nfX7l12pfRPpzZWCHTEkw','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483644','prod-end-to-end-og7nh','End-to-end School','-2147483644','e2e',NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'2P7yt0Ob9876kyCt/5bfl6OIt0/gzc6Q6UBznLXVvnGcrOgA+TMk5nmsSuNCOLkTbpyxXAzdLC/K
iaABA6TGtNT3x8smQBccbcYplHe8lLiRrtjB7JDipUwF71gEDBk3','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483645','prod-renseikai-8xr29','Renseikai School','-2147483645','renseikai',NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'LKUW9rlUG0vCvRxCvsT+sHY7yeU4airnm+UJP4nN0hVV98vpJ+5myUyC+awB/GUxRbQqTbxEE4pX
nVUPQUGIZjNyPGcpWPAZ70u2zloFVe77mLLWT+JydZiN956Om5E/','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483646','prod-synersia-trscc','Synersia School','-2147483646','synersia','https://storage.googleapis.com/prod-tokyo-backend/user-upload/multi-tenant-logo/synersia-logo.png',NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'3nSTYM0GFDXhUikExWyq0UHqto/eTU13BDscs4At4HeuunKYXyTvb4PYpW/2E8WGjcJc0ig20zr5
fS9iNGXEwiartESwse6fqbzY0Kgt8YTHFoTpKMtnDKJh9o33W7uf','Ke49jEokHXStXK0hgq/yOA==','UDpIUi480wpJQd81PViYkQ==','5GsVm8n90zZPH0nMdSuIHA=='),
	 ('-2147483647','-2147483647','JPREP School','-2147483647',NULL,NULL,NULL,'2022-05-10 11:10:23.413484+07','2022-05-10 11:10:23.413484+07',NULL,NULL,NULL,NULL,NULL),
	 ('-2147483648','manabie-b2c-ulfo0','Manabie School','-2147483648','manabie',NULL,'COUNTRY_VN','2022-05-10 11:10:23.413484+07','2022-05-10 16:26:45.355945+07',NULL,'','','','')
 ON CONFLICT DO NOTHING;
