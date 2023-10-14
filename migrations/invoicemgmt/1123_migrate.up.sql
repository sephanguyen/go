UPDATE invoice_schedule 
  SET scheduled_date = invoice_date + interval '24 hours';

ALTER TABLE invoice_schedule
  ALTER COLUMN scheduled_date SET NOT NULL;