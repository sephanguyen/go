@quarantined
Feature: Bulk Issue Invoice
  As an HQ manager or admin
  I can bulk change all "Draft" or "Failed" invoices to the status "Issued"

  Scenario Outline: Admin issues an invoice with default payment of student successfully
    Given there are existing invoices with "<status>" status
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" issues invoices in bulk with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "OK" status code
    And invoices status is updated to "ISSUED" status
    And invoices exported tag is set to "false"
    And payments exported tag is set to "false"
    And there are pending payment records for students created with payment method "<payment-method>" and "<due-date>" "<expiry-date>"
    And action log record for each invoice is recorded with "INVOICE_ISSUED" action log type

    Examples:
      | signed-in user | status | bulk-payment-method          | default-payment-method | payment-method    | due-date | expiry-date |
      | school admin   | DRAFT  | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     |
      | hq staff       | FAILED | BULK_ISSUE_DEFAULT_PAYMENT   | CONVENIENCE_STORE      | CONVENIENCE_STORE | TODAY+1  | TODAY+2     |
      | hq staff       | DRAFT  | BULK_ISSUE_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+2  | TODAY+3     |

  Scenario Outline: Admin failed to issue an invoice with non-existing invoice ID from the invoice IDs
    Given there are existing invoices with "<status>" status
    And these invoice for students have default payment method "<default-payment-method>"
    And one invoice ID is added to the request but is non-existing
    When "<signed-in user>" issues invoices in bulk with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "Internal" status code

    Examples:
      | signed-in user | status | bulk-payment-method          | default-payment-method | due-date | expiry-date |
      | school admin   | FAILED | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           | TODAY    | TODAY+1     |
      | hq staff       | DRAFT  | BULK_ISSUE_CONVENIENCE_STORE |                        | TODAY+1  | TODAY+2     |

  Scenario Outline: Admin failed to issue an invoice with invalid payment method
    Given there are existing invoices with "<status>" status
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" issues invoices in bulk with payment method "INVALID_PAYMENT_METHOD" and "<due-date>" "<expiry-date>"
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | status | default-payment-method | due-date | expiry-date |
      | school admin   | FAILED | DIRECT_DEBIT           | TODAY    | TODAY+1     |
      | hq staff       | DRAFT  | CONVENIENCE_STORE      | TODAY+1  | TODAY+2     |

  Scenario Outline: Admin failed to issue an invoice with negative invoice total amount
    Given there are existing invoices with "<status>" status
    And one invoice has negative total amount
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" issues invoices in bulk with payment method "<bulk-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | status | bulk-payment-method          | default-payment-method | due-date | expiry-date |
      | school admin   | FAILED | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           | TODAY    | TODAY+1     |
      | hq staff       | DRAFT  | BULK_ISSUE_DEFAULT_PAYMENT   | CONVENIENCE_STORE      | TODAY+1  | TODAY+2     |
      | hq staff       | DRAFT  | BULK_ISSUE_CONVENIENCE_STORE |                        | TODAY+2  | TODAY+3     |

  Scenario Outline: Admin failed to issue an invoice with due date after expiry
    Given there are existing invoices with "<status>" status
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" issues invoices in bulk with payment method "<bulk-payment-method>" and "<default-payment-method>" due date after expiry date
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | status | bulk-payment-method          | default-payment-method |
      | school admin   | FAILED | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           |
      | hq staff       | DRAFT  | BULK_ISSUE_DEFAULT_PAYMENT   | CONVENIENCE_STORE      |
      | hq staff       | DRAFT  | BULK_ISSUE_CONVENIENCE_STORE |                        |
