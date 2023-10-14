@quarantined
Feature: Bulk Payment Validation
  As an HQ manager or admin
  I can perform bulk payment validation

  # File content types
  # all-return-codes-transferred            -- all return codes are successful and indicate payments are transferred
  # mixed-return-codes                      -- combinations of successful and failed return codes
  # existing-return-codes                   -- payment records with existing return codes
  # duplicate-payment-records               -- duplicate payment records with same and different created date
  # void-invoice-no-status-change           -- voided invoice and failed payment status that has no status change
  # failed-invoice-no-status-change         -- convenience store with failed invoice and failed payment status that has no status change
  Background:
    Given unleash feature flag is "disable" with feature name "Invoice_InvoiceManagement_BackOffice_BulkAddAndValidatePayments"

  Scenario Outline: Admin validates convenience store payment file successfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "ISSUED" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "<payment-method>" payment method
    And has "<file-content-type>" file content type with payment date "<payment-date>" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "OK" status code
    And receives expected result with correct DB records based on "<file-content-type>" file content type

    Examples:
      | signed-in user | payment-method    | file-content-type     | existing-invoices | payment-date |
      | school admin   | CONVENIENCE_STORE | mixed-return-codes    | 13                | TODAY+1      |
      | hq staff       | CONVENIENCE_STORE | existing-return-codes | 4                 | TODAY+2      |

  Scenario Outline: Admin validates direct debit payment file successfully
    Given there are "12" preexisting number of existing invoices with "ISSUED" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "<payment-method>" payment method
    And has "<file-content-type>" file content type with payment date "<payment-date>" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "OK" status code
    And receives expected result with correct DB records based on "<file-content-type>" file content type

    Examples:
      | signed-in user | payment-method | file-content-type            | payment-date |
      | school admin   | DIRECT_DEBIT   | all-return-codes-transferred | TODAY+1      |
      | hq staff       | DIRECT_DEBIT   | mixed-return-codes           | TODAY+2      |

  Scenario Outline: Admin validates convenience store payment file with duplicate payments successfully
    Given there are "1" preexisting number of existing invoices with "ISSUED" status
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has duplicate payment records with "<created-date>" date and "<result-code>" result code sequence on payment file
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And receives expected record for duplicate payment with "<actual-result-code>" result code

    # Result code used:
    # CR0 -> C-R0 paid by customer and already transferred
    # CR1 -> C-R1 paid by customer but not transferred yet
    # CR2 -> C-R2 prompt report was cancelled
    Examples:
      | signed-in user | created-date            | result-code | actual-result-code |
      | school admin   | TODAY-TODAY-TODAY+1     | CR0-CR1-CR2 | CR2                |
      | school admin   | TODAY-TODAY+1-TODAY+1   | CR0-CR2-CR1 | CR1                |
      | school admin   | TODAY+1-TODAY-TODAY     | CR1-CR2-CR0 | CR1                |
      | hq staff       | TODAY+1-TODAY+1-TODAY   | CR1-CR0-CR2 | CR0                |
      | hq staff       | TODAY+2-TODAY-1-TODAY+1 | CR2-CR0-CR1 | CR2                |
      | hq staff       | TODAY+2-TODAY+3-TODAY+1 | CR2-CR0-CR1 | CR0                |

  Scenario Outline: Admin validate paid invoices and successful payments for both payment method unsuccessfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "PAID" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_SUCCESSFUL" status
    And there is an existing payment file for "<payment-method>" payment method
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-method    | existing-invoices |
      | school admin   | CONVENIENCE_STORE | 1                 |
      | hq staff       | DIRECT_DEBIT      | 2                 |

  Scenario Outline: Admin validate failed invoices and failed payments for both payment method unsuccessfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "FAILED" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_SUCCESSFUL" status
    And these existing payments have existing result code "<result-code>"
    And there is an existing payment file for "<payment-method>" payment method
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-method    | existing-invoices | result-code |
      | school admin   | CONVENIENCE_STORE | 1                 | D-R0        |
      | hq staff       | DIRECT_DEBIT      | 2                 | C-R0        |

  Scenario Outline: Admin validate file with void invoice and failed payment
    Given there are "<existing-invoices>" preexisting number of existing invoices with "VOID" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_FAILED" status
    And these existing payments have existing result code "<result-code>"
    And there is an existing payment file for "<payment-method>" payment method
    And has "<file-content-type>" file content type with payment date "<payment-date>" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "OK" status code
    And receives expected result with correct DB records based on "<file-content-type>" file content type

    Examples:
      | signed-in user | payment-method    | existing-invoices | file-content-type             | payment-date | result-code |
      | school admin   | CONVENIENCE_STORE | 4                 | void-invoice-no-status-change | TODAY+1      | C-R0        |
      | hq staff       | DIRECT_DEBIT      | 4                 | void-invoice-no-status-change | TODAY+1      | D-R0        |

  Scenario Outline: Admin validate file with void invoice and failed payment that has existing result code
    Given there are "<existing-invoices>" preexisting number of existing invoices with "VOID" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "<payment-method>" payment method
    And has "<file-content-type>" file content type with payment date "<payment-date>" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "<payment-method>" payment method
    Then receives "OK" status code
    And receives expected result with correct DB records based on "<file-content-type>" file content type

    Examples:
      | signed-in user | payment-method    | existing-invoices | file-content-type             | payment-date |
      | school admin   | CONVENIENCE_STORE | 4                 | void-invoice-no-status-change | TODAY+1      |
      | hq staff       | DIRECT_DEBIT      | 4                 | void-invoice-no-status-change | TODAY+1      |

  Scenario Outline: Admin validates convenience store payment file with failed invoice and failed payment
    Given there are "<existing-invoices>" preexisting number of existing invoices with "FAILED" status
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has "<file-content-type>" file content type with payment date "<payment-date>" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And receives expected result with correct DB records based on "<file-content-type>" file content type

    Examples:
      | signed-in user | existing-invoices | file-content-type               | payment-date |
      | school admin   | 4                 | failed-invoice-no-status-change | TODAY+1      |
      | hq staff       | 4                 | failed-invoice-no-status-change | TODAY+1      |

  Scenario Outline: Admin validates payment file that contains a payment that has different payment method
    Given there are "<existing-invoices>" preexisting number of existing invoices with "ISSUED" status
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "<file-payment-method>" payment method
    And has "mixed-return-codes" file content type with payment date "TODAY+1" for successful payments
    When "<signed-in user>" signed-in user uploads the payment file for "<file-payment-method>" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | existing-invoices | file-payment-method | payment-method    |
      | school admin   | 4                 | DIRECT_DEBIT        | CONVENIENCE_STORE |
      | hq staff       | 4                 | CONVENIENCE_STORE   | DIRECT_DEBIT      |
