/**
Set the outstanding_balance, amount_paid and amount_refunded columns of existing invoices

For outstanding_balance:
    - If the invoice status is PAID or REFUNDED, the value should be 0
    - If not, then the outstanding_balance should be equal to invoice total

For amount_paid:
    - If the status is PAID, the amount_paid should be equal to the invoice total
    - If not, the value should be 0

For amount_refunded:
    - If the status is REFUNDED, the amount_refunded should be equal to the invoice total
    - If not, the value should be 0
**/

UPDATE invoice i
SET 
	outstanding_balance = 
		CASE i.status
			WHEN 'PAID' THEN 0
			WHEN 'REFUNDED' THEN 0
		ELSE i.total
		END,
	amount_paid = 
		 CASE i.status
			WHEN 'PAID' THEN i.total
		 ELSE 0
		 END,
	amount_refunded = 
		 CASE i.status
			WHEN 'REFUNDED' THEN i.total
		 ELSE 0
		 END
WHERE amount_paid IS NULL 
	AND amount_refunded IS NULL
	AND outstanding_balance IS NULL;