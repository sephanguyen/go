-- Add column vendor_user_id for internal_admin_user
ALTER TABLE IF EXISTS ONLY public.internal_admin_user ADD COLUMN IF NOT EXISTS vendor_user_id TEXT NULL;

-- Remove current data
DELETE FROM public.internal_admin_user WHERE resource_path IS NOT NULL;

-- Fix resource path
-- (-2147483648) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSR7WHVHCQ5SCC62KE', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483648')
ON CONFLICT DO NOTHING;
        
-- (-2147483647) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSP5AEEXRCHTW84BHC', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483647')
ON CONFLICT DO NOTHING;
        
-- (-2147483646) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSZE0YMG5GS22RC9VZ', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483646')
ON CONFLICT DO NOTHING;
        
-- (2147483646) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSMEF22D0E1VVR6F47', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483646')
ON CONFLICT DO NOTHING;
        
-- (-2147483645) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSS6CRAAKG6SQTBYW7', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483645')
ON CONFLICT DO NOTHING;
        
-- (2147483645) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXSX305NWEPQWGDFRRC', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483645')
ON CONFLICT DO NOTHING;
        
-- (-2147483644) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXTTYFS0B1X3E0GW6JP', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483644')
ON CONFLICT DO NOTHING;
        
-- (-2147483643) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXTQKV7XJ5YTH4VPHZF', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483643')
ON CONFLICT DO NOTHING;
        
-- (2147483643) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXTGGPE27H5B24TR18Q', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483643')
ON CONFLICT DO NOTHING;
        
-- (-2147483642) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXVSZTQZR998AYJHVAH', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642')
ON CONFLICT DO NOTHING;
        
-- (-2147483641) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXVWW8Y9YYY0D1FJ43W', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483641')
ON CONFLICT DO NOTHING;
        
-- (2147483641) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXVDX5NRNY7Q3M78SCD', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483641')
ON CONFLICT DO NOTHING;
        
-- (-2147483640) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXVZG6V0WEYW98Q0J77', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483640')
ON CONFLICT DO NOTHING;
        
-- (-2147483639) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXWDT3T9KQ9EACN8QXZ', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483639')
ON CONFLICT DO NOTHING;
        
-- (-2147483638) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXWV6WHVA2ZFJPK5GB2', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483638')
ON CONFLICT DO NOTHING;
        
-- (-2147483637) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXWW1Y3A76ZAEND3GK5', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483637')
ON CONFLICT DO NOTHING;
        
-- (-2147483635) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXW6SCMP94S5V8J1HFD', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483635')
ON CONFLICT DO NOTHING;
        
-- (-2147483634) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXXD2HQZQ3KMC1BP83S', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483634')
ON CONFLICT DO NOTHING;
        
-- (-2147483631) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXXFDNMDGTT17PBTWJE', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483631')
ON CONFLICT DO NOTHING;
        
-- (2147483631) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXX0KXJZH6JD1SW8XP7', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483631')
ON CONFLICT DO NOTHING;
        
-- (-2147483630) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXXN7SFGC6PEHBJZA5G', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483630')
ON CONFLICT DO NOTHING;
        
-- (2147483630) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXYB74TSPXWWP2HPS7M', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483630')
ON CONFLICT DO NOTHING;
        
-- (-2147483629) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXYZEMWV781WZQMY8V1', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483629')
ON CONFLICT DO NOTHING;
        
-- (2147483629) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXYYBDNP1EMCNC0DMGD', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483629')
ON CONFLICT DO NOTHING;
        
-- (-2147483628) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXZMFRWT7JQNHDR723X', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483628')
ON CONFLICT DO NOTHING;
        
-- (-2147483627) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXZ67FJ79DVJR5DPAS5', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483627')
ON CONFLICT DO NOTHING;
        
-- (-2147483626) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCXZFDN4W67ZDRBC851Z', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483626')
ON CONFLICT DO NOTHING;
        
-- (2147483626) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY0W1SKFXH9YBR720DQ', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483626')
ON CONFLICT DO NOTHING;
        
-- (-2147483625) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY1ZGJQ7WS6WX3XGD4S', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483625')
ON CONFLICT DO NOTHING;
        
-- (2147483625) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY1XSK2KMVDQR2H7A7J', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483625')
ON CONFLICT DO NOTHING;
        
-- (-2147483624) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY1MRE9A48FJ2ADVQJS', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483624')
ON CONFLICT DO NOTHING;
        
-- (2147483624) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY192WDFXTDQ4S3Z0Q9', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '2147483624')
ON CONFLICT DO NOTHING;
        
-- (-2147483623) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY2Z7V05M55QB5KEZ4X', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483623')
ON CONFLICT DO NOTHING;
        
-- (-2147483622) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY22JCYVGTZAEQ2ESYN', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483622')
ON CONFLICT DO NOTHING;
        
-- (100013) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY2PG9Y54KRC1JZZ5XA', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '100013')
ON CONFLICT DO NOTHING;
        
-- (100012) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY20CGNV304V9M7Y9GS', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '100012')
ON CONFLICT DO NOTHING;
        
-- (100000) --
INSERT INTO public.internal_admin_user (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01H7FDSCY2RPT8XZ1P53D0X91T', true, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '100000')
ON CONFLICT DO NOTHING;

-- Update vendor_user_id for internal_admin_user
UPDATE internal_admin_user AS iau
    SET vendor_user_id = MD5(user_id);
