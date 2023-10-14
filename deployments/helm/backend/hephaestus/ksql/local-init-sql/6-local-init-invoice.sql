\connect invoicemgmt;


INSERT INTO public.students(student_id, current_grade, updated_at, created_at, resource_path) VALUES
    ('01GWHQY4BBRYZ7G0XDEQN22TN0', 1,  now(), now(), '-2147483642');

INSERT INTO public.invoice (invoice_id,invoice_sequence_number,type,status,student_id,sub_total,total,outstanding_balance,amount_paid,amount_refunded,resource_path,created_at,updated_at,deleted_at) VALUES
    ('01GWHQV538V3AP0WVDFMZ0TJ2W',1,'MANUAL','DRAFT','01GWHQY4BBRYZ7G0XDEQN22TN0',1000.00,1000.00,0.00,0.00,0.00,'-2147483642',now(),now(),NULL);
