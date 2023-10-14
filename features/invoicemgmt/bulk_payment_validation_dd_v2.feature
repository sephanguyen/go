@major
Feature: Bulk Payment Validation for Direct Debit Scenarios
  As an HQ manager or admin
  I can perform bulk payment validation

  # Scenarios and result code categories are based on https://manabie.atlassian.net/wiki/spaces/LT/pages/517308417/Payment+Result+Patterns Phase 2
  # Just set the feature flag to "enable" to test the bulk payment validation improvement
  Background:
    Given unleash feature flag is "enable" with feature name "Invoice_InvoiceManagement_BackOffice_BulkAddAndValidatePayments"
    And unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_ImproveBulkPaymentValidation"

  # Scenario 1
  Scenario: Admin validates direct debit payment file successfully
    Given there are "1" number of existing "ISSUED" invoices with total "500.00" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "0" on its file content
    And has "TODAY+1" payment date on the request
    When "school admin" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "OK" status code
    And payments have "D-R0" result code with correct expected result

  # Scenario 2
  Scenario Outline: Admin validates direct debit payment file failed result code [1 to 4,8,9]
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on the request
    When "<signed-in user>" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | amount | category | result-code |
      | school admin   | TODAY+1      | 200.00 | 1        | D-R1        |
      | hq staff       | TODAY+2      | 120.00 | 2        | D-R2        |
      | hq staff       | TODAY+2      | 150.00 | 3        | D-R3        |
      | hq staff       | TODAY+1      | 160.00 | 4        | D-R4        |
      | hq staff       | TODAY+1      | 300.00 | 8        | D-R8        |
      | school admin   | TODAY+3      | 330.00 | 9        | D-R9        |

  # Scenario 9
  Scenario Outline: Admin validates direct debit payment file failed cash amount
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on the request
    And has amount mismatched on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | amount | category | result-code |
      | hq staff       | TODAY+2      | 550.00 | 0        | D-R0-1      |
      | school admin   | TODAY+1      | 200.00 | 1        | D-R1-1      |
      | hq staff       | TODAY+2      | 222.00 | 2        | D-R2-1      |
      | hq staff       | TODAY+2      | 655.00 | 3        | D-R3-1      |
      | hq staff       | TODAY+1      | 5.00   | 4        | D-R4-1      |
      | hq staff       | TODAY+1      | 33.00  | 8        | D-R8-1      |
      | school admin   | TODAY+3      | 980.00 | 9        | D-R9-1      |

  # Scenario 11, 12
  Scenario: Admin validates direct debit payment file successfully void invoice failed payment
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "TODAY+1" payment date on the request
    When "school admin" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | amount | category | result-code |
      | hq staff       | TODAY+2      | 777.00 | 0        | D-R0-2      |
      | school admin   | TODAY+1      | 800.00 | 1        | D-R1-2      |
      | hq staff       | TODAY+2      | 966.00 | 2        | D-R2-2      |
      | hq staff       | TODAY+2      | 669.00 | 3        | D-R3-2      |
      | hq staff       | TODAY+1      | 25.00  | 4        | D-R4-2      |
      | hq staff       | TODAY+1      | 233.00 | 8        | D-R8-2      |
      | school admin   | TODAY+3      | 981.00 | 9        | D-R9-2      |

  # Scenario 13 amount not match
  Scenario: Admin validates direct debit payment file successfully void invoice failed payment and incorrect cash amount
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_FAILED" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "TODAY+1" payment date on the request
    And has amount mismatched on its file content
    When "school admin" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "OK" status code
    And payments have "<result-code>" result code with correct expected result

    Examples:
      | signed-in user | payment-date | amount | category | result-code |
      | hq staff       | TODAY+2      | 778.00 | 0        | D-R0-3      |
      | school admin   | TODAY+1      | 801.00 | 1        | D-R1-3      |
      | hq staff       | TODAY+2      | 967.00 | 2        | D-R2-3      |
      | hq staff       | TODAY+2      | 670.00 | 3        | D-R3-3      |
      | hq staff       | TODAY+1      | 26.00  | 4        | D-R4-3      |
      | hq staff       | TODAY+1      | 234.00 | 8        | D-R8-3      |
      | school admin   | TODAY+3      | 982.00 | 9        | D-R9-3      |

  # Scenario 17-A paid invoice and success payment
  Scenario: Admin validate paid invoices and successful payments for both payment method unsuccessfully
    Given there are "1" number of existing "PAID" invoices with total "50.00" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_SUCCESSFUL" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    When "school admin" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code

  # Scenario 17-B void invoice and failed payment existing result code
  Scenario Outline: Admin validates direct debit payment file w/ void invoice status and payment failed
    Given there are "1" number of existing "VOID" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And these existing payments have existing result code "<existing-result-code>"
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-date | amount | category | existing-result-code |
      | hq staff       | TODAY+2      | 778.00 | 0        | D-R0                 |
      | school admin   | TODAY+1      | 801.00 | 1        | D-R1                 |
      | hq staff       | TODAY+2      | 967.00 | 2        | D-R2                 |
      | hq staff       | TODAY+2      | 670.00 | 3        | D-R3                 |
      | hq staff       | TODAY+1      | 26.00  | 4        | D-R4                 |
      | hq staff       | TODAY+1      | 234.00 | 8        | D-R8                 |
      | school admin   | TODAY+3      | 982.00 | 9        | D-R9                 |

  # Scenario 17-C issued invoice and failed payment existing result code
  Scenario Outline: Admin validates direct debit payment file w/ issued invoice status and payment failed
    Given there are "1" number of existing "ISSUED" invoices with total "<amount>" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And these existing payments have existing result code "<existing-result-code>"
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "<category>" on its file content
    And has "<payment-date>" payment date on its file content
    When "<signed-in user>" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-date | amount | category | existing-result-code |
      | hq staff       | TODAY+2      | 779.00 | 0        | D-R0                 |
      | school admin   | TODAY+1      | 802.00 | 1        | D-R1                 |
      | hq staff       | TODAY+2      | 968.00 | 2        | D-R2                 |
      | hq staff       | TODAY+2      | 671.00 | 3        | D-R3                 |
      | hq staff       | TODAY+1      | 27.00  | 4        | D-R4                 |
      | hq staff       | TODAY+1      | 235.00 | 8        | D-R8                 |
      | school admin   | TODAY+3      | 983.00 | 9        | D-R9                 |

  # Scenario 19, 24
  Scenario: Admin validates direct debit payment file that contains a payment that has different payment method
    Given there are "1" number of existing "ISSUED" invoices with total "241.00" amount
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "02" on its file content
    And has "TODAY+1" payment date on its file content
    When "hq staff" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code

  # Scenario 20
  Scenario Outline: Admin validates direct debit payment file that contains an invalid category
    Given there are "1" number of existing "ISSUED" invoices with total "241.00" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "-1" on its file content
    And has "TODAY+1" payment date on its file content
    When "hq staff" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code

  # Scenario 22
  Scenario: Admin validates direct debit payment file with non existing payment record
    Given there are "1" number of existing "ISSUED" invoices with total "500.00" amount
    And there are existing payments for those invoices for "DIRECT_DEBIT" payment method with "PAYMENT_PENDING" status
    And there is an existing payment file for "DIRECT_DEBIT" payment method
    And has result code category "0" on its file content
    And has "TODAY+1" payment date on its file content
    And there is a payment that is not match in our system
    When "school admin" signed-in user uploads the payment file for "DIRECT_DEBIT" payment method
    Then receives "InvalidArgument" status code
