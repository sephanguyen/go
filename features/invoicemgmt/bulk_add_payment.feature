Feature: Bulk Add Payment
  As an HQ manager or admin
  I can bulk add payment for invoices that has invoice status with failed or no existing payments 

  @major
  Scenario Outline: Admin bulk add an invoices with single invoice type and failed latest payment status successfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "ISSUED" status
    And these invoices has "<invoice-type>" type 
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_FAILED" status
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" bulk add payment for these invoices with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "OK" status code
    And payments exported tag is set to "false"
    And there are pending payment records for students created with payment method "<payment-method>" and "<due-date>" "<expiry-date>"
    And bulk payment record is created successfully with payment method "<bulk-payment-method>"
    And action log record for each invoice is recorded with "PAYMENT_ADDED" action log type

    Examples:
      | signed-in user | existing-invoices | bulk-payment-method            | default-payment-method | payment-method    | due-date | expiry-date | invoice-type |
      | school admin   | 1                 | BULK_PAYMENT_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     | SCHEDULED    | 
      | hq staff       | 2                 | BULK_PAYMENT_DEFAULT_PAYMENT   | CONVENIENCE_STORE      | CONVENIENCE_STORE | TODAY+1  | TODAY+2     | MANUAL       |
      | hq staff       | 1                 | BULK_PAYMENT_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+2  | TODAY+3     | MANUAL       |
  
  @major
  Scenario Outline: Admin bulk add an invoices with multiple invoice type and latest payment status successfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "ISSUED" status
    And these invoices has "SCHEDULED" type 
    And there are existing payments for those invoices for "<payment-method>" payment method with "PAYMENT_FAILED" status
    And another "<existing-invoices>" preexisting number of existing invoices with "ISSUED" status
    And these invoices has "MANUAL" type
    And there are no payments for these invoices
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" bulk add payment for these invoices with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "OK" status code
    And payments exported tag is set to "false"
    And there are pending payment records for students created with payment method "<payment-method>" and "<due-date>" "<expiry-date>"
    And bulk payment record is created successfully with payment method "<bulk-payment-method>"
    And action log record for each invoice is recorded with "PAYMENT_ADDED" action log type

    Examples:
      | signed-in user | existing-invoices | bulk-payment-method            | default-payment-method | payment-method    | due-date | expiry-date | 
      | school admin   | 1                 | BULK_PAYMENT_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     |
      | hq staff       | 2                 | BULK_PAYMENT_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+1  | TODAY+2     |

  
  Scenario Outline: Admin bulk add an invoices with existing payments that are not failed
    Given there are "1" preexisting number of existing invoices with "ISSUED" status
    And these invoices has "<invoice-type>" type 
    And there are existing payments for those invoices for "<payment-method>" payment method with "<payment-status>" status
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" bulk add payment for these invoices with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | payment-status     | bulk-payment-method            | default-payment-method | payment-method    | due-date | expiry-date | invoice-type |
      | school admin   | PAYMENT_PENDING    | BULK_PAYMENT_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     | SCHEDULED    |
      | hq staff       | PAYMENT_SUCCESSFUL | BULK_PAYMENT_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+2  | TODAY+3     | MANUAL       |

  Scenario Outline: Admin bulk add an invoices with non existing invoice on the system
    Given there are "1" preexisting number of existing invoices with "ISSUED" status
    And these invoices has "SCHEDULED" type 
    And there are existing payments for those invoices for "CONVENIENCE_STORE" payment method with "PAYMENT_FAILED" status
    And these invoice for students have default payment method "CONVENIENCE_STORE"
    And one invoice ID is added to the request but is non-existing
    When "school admin" bulk add payment for these invoices with payment method "BULK_PAYMENT_CONVENIENCE_STORE" and "TODAY" "TODAY+1"
    Then receives "Internal" status code

