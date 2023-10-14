@major
Feature: Bulk Payment Validation
  As an HQ manager or admin
  I can perform bulk payment validation

  # Scenarios and result code categories are based on https://manabie.atlassian.net/wiki/spaces/LT/pages/517308417/Payment+Result+Patterns Phase 2
  # Just set the feature flag to "enable" to test the bulk payment validation improvement
  Background:
    Given unleash feature flag is "enable" with feature name "Invoice_InvoiceManagement_BackOffice_BulkAddAndValidatePayments"
    And unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_ImproveBulkPaymentValidation"

  # Scenario 3 CS
  Scenario: Admin validates convenience store payment file successfully
    Given there are "1" number of existing "ISSUED" invoices with total "500.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "02" on its file content
    And has "TODAY+1" payment date on its file content
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R0" result code with correct expected result

  # Scenario 4 CS
  Scenario: Admin validates convenience store payment file paid by customer but not transferred
    Given there are "1" number of existing "ISSUED" invoices with total "84.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "01" on its file content
    When "hq staff" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R1" result code with correct expected result

  # Scenario 5 CS
  Scenario: Admin validates convenience store payment file report cancelled
    Given there are "1" number of existing "ISSUED" invoices with total "90.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "03" on its file content
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R2" result code with correct expected result

  # Scenario 6 CS
  Scenario: Admin validates convenience store payment file report cancelled and has existing result code not transferred
    Given there are "1" number of existing "ISSUED" invoices with total "65.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And these existing payments have existing result code "C-R1"
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "03" on its file content
    When "hq staff" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R2" result code with correct expected result

  # Scenario 7 CS
  Scenario: Admin validates convenience store payment file not transferred and has existing result code report cancelled
    Given there are "1" number of existing "ISSUED" invoices with total "33.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And these existing payments have existing result code "C-R2"
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "01" on its file content
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R1" result code with correct expected result

  # Scenario 8 CS
  Scenario Outline: Admin validates convenience store payment file successfully with existing result code
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And these existing payments have existing result code "<existing-result-code>"
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "02" on its file content
    And has "<payment-date>" payment date on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R0" result code with correct expected result

    Examples:
      | signed-in user | payment-date | existing-result-code | amount |
      | school admin   | TODAY+1      | C-R1                 | 200.00 |
      | hq staff       | TODAY+2      | C-R2                 | 100.00 |

  # Scenario 10 CS
  Scenario Outline: Admin validates convenience store payment file failed category 01-03
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    And has amount mismatched on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | category | result-code | amount |
      | school admin   | TODAY+1      | 01       | C-R1-1      | 800.00 |
      | hq staff       | TODAY+2      | 02       | C-R0-1      | 400.00 |
      | hq staff       | TODAY+2      | 03       | C-R2-1      | 300.00 |

  # Scenario 11-12 CS
  Scenario Outline: Admin validates convenience store payment file w/ void invoice status and payment failed
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | category | result-code | amount |
      | school admin   | TODAY+1      | 01       | C-R1-2      | 150.00 |
      | hq staff       | TODAY+2      | 02       | C-R0-2      | 50.00  |
      | hq staff       | TODAY+2      | 03       | C-R2-2      | 950.00 |

  # Scenario 13 CS
  Scenario Outline: Admin validates convenience store payment file w/ void invoice status, payment failed and amount mismatched
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    And has amount mismatched on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | category | result-code | amount |
      | school admin   | TODAY+1      | 01       | C-R1-3      | 10.00  |
      | hq staff       | TODAY+2      | 02       | C-R0-3      | 700.00 |
      | hq staff       | TODAY+2      | 03       | C-R2-3      | 240.00 |

  # Scenario 14 category 01 and 03 CS
  Scenario Outline: Admin validates convenience store payment file w/ amount mismatched and report cancelled
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | category | result-code | amount |
      | school admin   | TODAY+1      | 01       | C-R1-2      | 40.00  |
      | hq staff       | TODAY+2      | 03       | C-R2-2      | 90.00  |

  # Scenario 15 category 02 CS
  Scenario Outline: Admin validates convenience store payment file w/ paid transferred
    Given there are "1" number of existing "ISSUED" invoices with total "44.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "02" on its file content
    And has "TODAY+2" payment date on its file content
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "C-R0-2" result code with correct expected result

  # Scenario 16 CS
  Scenario Outline: Admin validates convenience store payment file w/ issued invoice status, payment failed and amount mismatched
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    And has amount mismatched on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | category | result-code | amount |
      | school admin   | TODAY+1      | 01       | C-R1-3      | 80.00  |
      | hq staff       | TODAY+2      | 02       | C-R0-3      | 99.00  |
      | hq staff       | TODAY+2      | 03       | C-R2-3      | 26.00  |

  # Scenario 17-A CS paid invoice and success payment
  Scenario: Admin validate paid invoices and successful payments for both payment method unsuccessfully
    Given there are "1" number of existing "PAID" invoices with total "50.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_SUCCESSFUL" status
    And there is an existing payment file for "CONVENIENCE_STORE>" payment method
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE>" payment method
    Then receives "InvalidArgument" status code

  # Scenario 17-B CS void invoice and failed payment existing result code
  Scenario Outline: Admin validates convenience store payment file w/ void invoice status and payment failed
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And these existing payments have existing result code "<existing-result-code>"
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-date | category | result-code | amount | existing-result-code |
      | school admin   | TODAY+1      | 01       | C-R1-2      | 150.00 | C-R1                 |
      | hq staff       | TODAY+2      | 02       | C-R0-2      | 50.00  | C-R0                 |
      | hq staff       | TODAY+2      | 03       | C-R2-2      | 950.00 | C-R2                 |

  # Scenario 17-C CS issued invoice and failed payment existing result code
  Scenario Outline: Admin validates convenience store payment file w/ amount mismatched and report cancelled
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And these existing payments have existing result code "<existing-result-code>"
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "<category>" on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-date | category | result-code | amount | existing-result-code |
      | school admin   | TODAY+1      | 01       | C-R1-2      | 40.00  | C-R1                 |
      | hq staff       | TODAY+2      | 03       | C-R2-2      | 90.00  | C-R0                 |
      | hq staff       | TODAY+2      | 02       | C-R0-2      | 95.00  | C-R2                 |

  # Scenario 18 CS duplicates payment
  Scenario Outline: Admin validates convenience store payment file with duplicate payments successfully
    Given there are "1" number of existing "ISSUED" invoices with total "200.00" amount
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
      | signed-in user | created-date            | result-code    | actual-result-code |
      | school admin   | TODAY-TODAY-TODAY+1     | C-R0/C-R1/C-R2 | C-R2               |
      | school admin   | TODAY-TODAY+1-TODAY+1   | C-R0/C-R2/C-R1 | C-R1               |
      | school admin   | TODAY+1-TODAY-TODAY     | C-R1/C-R2/C-R0 | C-R1               |
      | hq staff       | TODAY+1-TODAY+1-TODAY   | C-R1/C-R0/C-R2 | C-R0               |
      | hq staff       | TODAY+2-TODAY-1-TODAY+1 | C-R2/C-R0/C-R1 | C-R2               |
      | hq staff       | TODAY+2-TODAY+3-TODAY+1 | C-R2/C-R0/C-R1 | C-R0               |

  # Scenario 19, 21, 24 CS
  Scenario: Admin validates payment file that contains a payment that has different payment method
    Given there are "1" number of existing "ISSUED" invoices with total "240.00" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "02" on its file content
    And has "TODAY+1" payment date on its file content
    When "hq staff" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "InvalidArgument" status code

  # Scenario 23 CS
  Scenario: Admin validates convenience store payment file with non existing payment record
    Given there are "1" number of existing "ISSUED" invoices with total "500.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "CONVENIENCE_STORE" payment method
    And has result code category "02" on its file content
    And has "TODAY+1" payment date on its file content
    And there is a payment that is not match in our system
    When "school admin" signed-in user uploads the payment file for "CONVENIENCE_STORE" payment method
    Then receives "InvalidArgument" status code
