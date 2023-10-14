ALTER TABLE public.order ADD COLUMN version_number int;

CREATE OR REPLACE FUNCTION fill_version_number() RETURNS TRIGGER
AS $$
    BEGIN       
        NEW.version_number = OLD.version_number + 1;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_order_version BEFORE UPDATE ON public.order FOR EACH ROW EXECUTE PROCEDURE fill_version_number();
