DROP TRIGGER IF EXISTS fill_in_order_version
ON public.order;

UPDATE public.order SET version_number = 0 WHERE version_number is null;
UPDATE public.student_product SET version_number = 0 WHERE version_number is null;

ALTER TABLE public.student_product ALTER COLUMN version_number SET NOT NULL;
ALTER TABLE public.order ALTER COLUMN version_number SET NOT NULL;

CREATE TRIGGER fill_in_student_product_version BEFORE UPDATE ON public.student_product FOR EACH ROW EXECUTE PROCEDURE fill_version_number();
CREATE TRIGGER fill_in_order_version BEFORE UPDATE ON public.order FOR EACH ROW EXECUTE PROCEDURE fill_version_number();
