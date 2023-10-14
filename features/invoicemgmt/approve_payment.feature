@quarantined
Feature: Approve Payment
  As an HQ manager or admin
  I am able to approve a payment
  Background:
    Given there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database

  Scenario Outline: Admin approves a payment successfully
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    And bill item has "pending" previous status
    And bill item has "<final-price-value>" final price value
    And there is "<payment-history>" payment history
    When admin approves payment with "<remarks>" remarks
    Then receives "OK" status code
    And invoice has "<new-invoice-status>" invoice status
    And invoice outstanding_balance set to "0.00"
    And invoice amount_paid set to "<amount-paid>"
    And invoice amount_refunded set to "<amount-refunded>"
    And action log record is recorded with "<action-log-type>" action log type
    And latest payment record has "<payment-status>" payment status

    Examples:
      | signed-in user | invoice-status | final-price-value | payment-history | remarks | new-invoice-status | action-log-type  | payment-status | amount-paid | amount-refunded |
      | school admin   | ISSUED         | 1000.00           | pending         | any     | PAID               | INVOICE_PAID     | SUCCESSFUL     | 1000.00     | 0.00            |
      | school admin   | ISSUED         | 500.53            | failed-pending  | none    | PAID               | INVOICE_PAID     | SUCCESSFUL     | 500.53      | 0.00            |
      | school admin   | ISSUED         | -100.99           | pending         | any     | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -100.99         |
      | school admin   | ISSUED         | -1.00             | failed-pending  | none    | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -1.00           |
      | hq staff       | ISSUED         | 200               | pending         | any     | PAID               | INVOICE_PAID     | SUCCESSFUL     | 200.00      | 0.00            |
      | hq staff       | ISSUED         | 100               | failed-pending  | none    | PAID               | INVOICE_PAID     | SUCCESSFUL     | 100.00      | 0.00            |
      | hq staff       | ISSUED         | -99.99            | pending         | any     | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -99.99          |
      | hq staff       | ISSUED         | -777.00           | failed-pending  | none    | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -777.00         |

  Scenario Outline: Admin failed to approve a payment with an invalid invoice status
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    When admin approves payment with "any" remarks
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invoice-status |
      | school admin   | PAID           |
      | school admin   | VOID           |
      | school admin   | REFUNDED       |
      | school admin   | DRAFT          |
      | school admin   | FAILED         |
      | hq staff       | PAID           |
      | hq staff       | VOID           |
      | hq staff       | REFUNDED       |
      | hq staff       | DRAFT          |
      | hq staff       | FAILED         |

  Scenario Outline: Admin failed to approve a payment with a non-existing invoice
    Given "<signed-in user>" logins to backoffice app
    And invoice ID is non-existing
    When admin approves payment with "any" remarks
    Then receives "Internal" status code

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |

  Scenario Outline: Admin approves a payment successfully with adjustment price on billing item
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    And bill item has "pending" previous status
    And bill item has "<final-price-value>" final price value
    And bill item has "<adjustment-price-value>" adjustment price value
    And there is "<payment-history>" payment history
    When admin approves payment with "<remarks>" remarks
    Then receives "OK" status code
    And invoice has "<new-invoice-status>" invoice status
    And invoice outstanding_balance set to "0.00"
    And invoice amount_paid set to "<amount-paid>"
    And invoice amount_refunded set to "<amount-refunded>"
    And action log record is recorded with "<action-log-type>" action log type
    And latest payment record has "<payment-status>" payment status

    Examples:
      | signed-in user | invoice-status | final-price-value | payment-history | remarks | new-invoice-status | action-log-type  | payment-status | adjustment-price-value | amount-paid | amount-refunded |
      | school admin   | ISSUED         | 1000.00           | PENDING         | any     | PAID               | INVOICE_PAID     | SUCCESSFUL     | 500                    | 500.00      | 0.00            |
      | hq staff       | ISSUED         | 100               | FAILED-PENDING  | none    | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | -500                   | 0.00        | -500.00         |
      | hq staff       | ISSUED         | -757.00           | FAILED-PENDING  | none    | PAID               | INVOICE_PAID     | SUCCESSFUL     | 600                    | 600.00      | 0.00            |
      | hq staff       | ISSUED         | -100              | FAILED-PENDING  | none    | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | -50                    | 0.00        | -50.00          |

  Scenario Outline: Admin approves a payment successfully with invoice adjustment
    Given "<signed-in user>" logins to backoffice app
    And there are "1" preexisting number of existing invoices with "DRAFT" status
    And this invoice has "<invoice-total>" total amount
    And admin adds "<record-count>" invoice adjustment with "<adjust-amount-to-create>" amount
    And apply the adjustment on the invoice
    And "<signed-in user>" issues invoice with "CASH" payment method
    And there is "<payment-history>" payment history
    When admin approves payment with "<remarks>" remarks
    Then receives "OK" status code
    And invoice has "<new-invoice-status>" invoice status
    And invoice outstanding_balance set to "0.00"
    And invoice amount_paid set to "<amount-paid>"
    And invoice amount_refunded set to "<amount-refunded>"
    And action log record is recorded with "<action-log-type>" action log type
    And latest payment record has "<payment-status>" payment status

    Examples:
      | signed-in user | invoice-total | payment-history | remarks | new-invoice-status | action-log-type  | payment-status | amount-paid | amount-refunded | record-count | adjust-amount-to-create |
      | school admin   | 100.00        | PENDING         | any     | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -200.00         | 1            | -300                    |
      | hq staff       | -100.00       | FAILED-PENDING  | none    | PAID               | INVOICE_PAID     | SUCCESSFUL     | 400.00      | 0.00            | 2            | 200&300                 |
      | hq staff       | 100.00        | FAILED-PENDING  | none    | PAID               | INVOICE_PAID     | SUCCESSFUL     | 600.00      | 0.00            | 1            | 500                     |
      | hq staff       | 100.00        | PENDING         | any     | REFUNDED           | INVOICE_REFUNDED | SUCCESSFUL     | 0.00        | -100.00         | 1            | -200                    |
