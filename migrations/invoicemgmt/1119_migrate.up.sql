ALTER PUBLICATION alloydb_publication ADD TABLE 
public.payment,
public.invoice_bill_item,
public.student_payment_detail,
public.student_payment_detail_action_log;

ALTER TABLE IF EXISTS public.student_payment_detail_action_log
    ALTER action_detail SET DEFAULT '{}'::jsonb;
