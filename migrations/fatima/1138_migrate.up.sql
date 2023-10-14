ALTER TABLE public.student_product ADD COLUMN version_number int default 0;

ALTER TABLE public.order
ALTER COLUMN version_number SET DEFAULT 0;
CREATE TRIGGER fill_in_order_version BEFORE UPDATE ON public.order FOR EACH ROW EXECUTE PROCEDURE fill_version_number();