UPDATE public.payment p
SET validated_date = pv.validation_date, updated_at = now()
FROM (
    SELECT pv.validation_date, pvd.payment_id
    FROM bulk_payment_validations_detail pvd
    INNER JOIN bulk_payment_validations pv
        ON pvd.bulk_payment_validations_id = pv.bulk_payment_validations_id
) pv
WHERE
    pv.payment_id = p.payment_id
    AND p.validated_date IS null
    AND p.result_code = 'C-R0'
    AND p.payment_status = 'PAYMENT_SUCCESSFUL'
    AND p.payment_method = 'CONVENIENCE_STORE';