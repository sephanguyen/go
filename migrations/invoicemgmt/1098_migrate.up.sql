-- --migration data for prefecture_code on Billing Address
DO $$DECLARE
    pref_code text;
    pref_id text;
BEGIN
  -- check chapter_ids exist to run second migration
    IF EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'billing_address' AND COLUMN_NAME = 'prefecture_id')
    THEN
    FOR pref_code, pref_id IN
        SELECT p.prefecture_code, p.prefecture_id
        FROM prefecture p
        JOIN billing_address ba ON p.prefecture_id = ba.prefecture_id
    loop
        UPDATE billing_address SET prefecture_code = pref_code WHERE prefecture_id = pref_id;
    END LOOP;
    END IF;
END$$;