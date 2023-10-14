DO $$
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.triggers 
        WHERE event_object_table = 'payment' 
          AND trigger_name = 'fill_in_payment_seq'
    ) THEN
        ALTER TABLE payment DISABLE TRIGGER fill_in_payment_seq;
    END IF;
END $$;