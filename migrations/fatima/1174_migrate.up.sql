-- Need to add this everytime add new org

--------------------------------------------------------------
--------------------- for public.order -----------------------
--------------------------------------------------------------

-- Drop the default value constraint on the column
ALTER TABLE public.order ALTER COLUMN order_sequence_number DROP DEFAULT;

-- Drop the sequence only if it exists
DROP SEQUENCE IF EXISTS public.order_sequence_number_seq CASCADE;

-- Loop through each distinct resource_path value in the organizations table and create a separate order sequence for each one
-- Need to add this everytime add new org
DO $$
DECLARE
    rp TEXT;
    orgs TEXT[] := ARRAY[
         '100000',
         '100012',
         '-2147483622',
         '-2147483623',
         '-2147483624',
         '-2147483625',
         '-2147483626',
         '-2147483627',
         '-2147483628',
         '-2147483629',
         '-2147483630',
         '-2147483631',
         '-2147483632',
         '-2147483633',
         '-2147483634',
         '-2147483635',
         '-2147483636',
         '-2147483637',
         '-2147483638',
         '-2147483639',
         '-2147483640',
         '-2147483641',
         '-2147483642',
         '-2147483643',
         '-2147483644',
         '-2147483645',
         '-2147483646',
         '-2147483647',
         '-2147483648'];
BEGIN
    FOR i IN 1..array_upper(orgs, 1) LOOP
        rp := orgs[i];
        EXECUTE format('CREATE SEQUENCE IF NOT EXISTS seq_order_%s START WITH %s', replace(rp, '-', '_'), COALESCE((SELECT MAX(order_sequence_number) + 1 FROM public.order WHERE resource_path = rp), 1));
        EXECUTE format('ALTER SEQUENCE seq_order_%s OWNED BY public.order.order_sequence_number', replace(rp, '-', '_'));    
    END LOOP;
END;
$$;

-- Update the fill_seq_order function to use separate order sequences for each resource_path
CREATE OR REPLACE FUNCTION fill_seq_order()
RETURNS TRIGGER AS $$
BEGIN
    NEW.order_sequence_number := nextval((SELECT CONCAT('seq_order_', replace(NEW.resource_path, '-', '_'))));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate the trigger to use the updated fill_seq_order function, also rename fill_in_order_seq to fill_seq_order_trigger
DROP TRIGGER IF EXISTS fill_in_order_seq ON public.order;
DROP TRIGGER IF EXISTS fill_seq_order_trigger ON public.order;
CREATE TRIGGER fill_seq_order_trigger
BEFORE INSERT ON public.order
FOR EACH ROW
EXECUTE FUNCTION fill_seq_order();


--------------------------------------------------------------
--------------------- for public.bill_item -------------------
--------------------------------------------------------------

-- Loop through each distinct resource_path value in the organizations table and create a separate bill_item sequence for each one
-- Need to add this everytime add new org
DO $$
DECLARE
    rp TEXT;
    orgs TEXT[] := ARRAY[
         '100000',
         '100012',
         '-2147483622',
         '-2147483623',
         '-2147483624',
         '-2147483625',
         '-2147483626',
         '-2147483627',
         '-2147483628',
         '-2147483629',
         '-2147483630',
         '-2147483631',
         '-2147483632',
         '-2147483633',
         '-2147483634',
         '-2147483635',
         '-2147483636',
         '-2147483637',
         '-2147483638',
         '-2147483639',
         '-2147483640',
         '-2147483641',
         '-2147483642',
         '-2147483643',
         '-2147483644',
         '-2147483645',
         '-2147483646',
         '-2147483647',
         '-2147483648'];
BEGIN
    FOR i IN 1..array_upper(orgs, 1) LOOP
        rp := orgs[i];
        EXECUTE format('CREATE SEQUENCE IF NOT EXISTS seq_bill_item_%s START WITH %s', replace(rp, '-', '_'), COALESCE((SELECT MAX(bill_item_sequence_number) + 1 FROM public.bill_item WHERE resource_path = rp), 1));
        EXECUTE format('ALTER SEQUENCE seq_bill_item_%s OWNED BY public.bill_item.bill_item_sequence_number', replace(rp, '-', '_'));
    END LOOP;
END;
$$;

-- Update the fill_seq_bill_item function to use separate bill_item sequences for each resource_path
CREATE OR REPLACE FUNCTION fill_seq_bill_item()
RETURNS TRIGGER AS $$
BEGIN
    NEW.bill_item_sequence_number := nextval((SELECT CONCAT('seq_bill_item_', replace(NEW.resource_path, '-', '_'))));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate the trigger to use the updated fill_seq_bill_item function, also rename fill_in_bill_item_seq to fill_seq_bill_item_trigger
DROP TRIGGER IF EXISTS fill_in_bill_item_seq ON public.bill_item;
DROP TRIGGER IF EXISTS fill_seq_bill_item_trigger ON public.bill_item;
CREATE TRIGGER fill_seq_bill_item_trigger
BEFORE INSERT ON public.bill_item
FOR EACH ROW
EXECUTE FUNCTION fill_seq_bill_item();