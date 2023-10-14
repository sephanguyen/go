-- Update the invoice_date and scheduled_date to hour 15 (00:00 JST) for scheduled invoice that uses hour 17 (00:00 VNT)
UPDATE invoice_schedule 
SET scheduled_date = scheduled_date AT TIME ZONE 'UTC' - interval '2 hours', 
    invoice_date = invoice_date AT TIME ZONE 'UTC' - interval '2 hours'
WHERE extract(hour from scheduled_date AT TIME ZONE 'UTC') = 17 
    AND extract(hour from invoice_date AT TIME ZONE 'UTC') = 17
    AND status = 'INVOICE_SCHEDULE_SCHEDULED';