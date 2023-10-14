@major
Feature: Bulk Issue Invoice V2 Improved
  As an HQ manager or admin
  I can bulk change all "Draft" invoices to the status "Issued"

  Background:
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_ImproveBulkIssueInvoice"

  Scenario Outline: Admin bulk issue invoices successfully
    Given there are "<existing-invoices>" preexisting number of existing invoices with "DRAFT" status
    And these invoices has "<invoice-type>" type
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" bulk issue invoices using v2 endpoint with payment method "<bulk-issue-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "OK" status code
    And invoices status is updated to "ISSUED" status
    And payments exported tag is set to "false"
    And there are pending payment records for students created with payment method "<payment-method>" and "<due-date>" "<expiry-date>"
    And action log record for each invoice is recorded with "INVOICE_BULK_ISSUED" action log type
    And bulk payment record is created successfully with payment method "<bulk-payment-method>"

    Examples:
      | signed-in user | existing-invoices | bulk-issue-payment-method    | default-payment-method | payment-method    | due-date | expiry-date | invoice-type | bulk-payment-method            |
      | school admin   | 10                | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     | SCHEDULED    | BULK_PAYMENT_DEFAULT_PAYMENT   |
      | hq staff       | 20                | BULK_ISSUE_DEFAULT_PAYMENT   | CONVENIENCE_STORE      | CONVENIENCE_STORE | TODAY+1  | TODAY+2     | MANUAL       | BULK_PAYMENT_DEFAULT_PAYMENT   |
      | hq staff       | 10                | BULK_ISSUE_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+2  | TODAY+3     | MANUAL       | BULK_PAYMENT_CONVENIENCE_STORE |

  Scenario: Admin failed to issue an invoice with negative invoice total amount
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And these invoices has "MANUAL" type
    And these invoice for students have default payment method "CONVENIENCE_STORE"
    And one invoice has negative total amount
    When "school admin" bulk issue invoices using v2 endpoint with payment method "BULK_ISSUE_CONVENIENCE_STORE" and "TODAY" "TODAY+1"
    Then receives "InvalidArgument" status code

  Scenario: Admin failed to issue an invoice with zero invoice total amount
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And these invoices has "MANUAL" type
    And these invoice for students have default payment method "CONVENIENCE_STORE"
    And one invoice has zero total amount
    When "school admin" bulk issue invoices using v2 endpoint with payment method "BULK_ISSUE_CONVENIENCE_STORE" and "TODAY" "TODAY+1"
    Then receives "InvalidArgument" status code

  Scenario Outline: Admin failed to issue an invoice not in draft status
    Given there are "1" preexisting number of existing invoices with "<invoice-status>" status
    And these invoices has "MANUAL" type
    And these invoice for students have default payment method "CONVENIENCE_STORE"
    When "school admin" bulk issue invoices using v2 endpoint with payment method "BULK_ISSUE_CONVENIENCE_STORE" and "TODAY" "TODAY+1"
    Then receives "InvalidArgument" status code

    Examples:
      | invoice-status |
      | ISSUED         |
      | PAID           |
      | FAILED         |
      | VOID           |

  Scenario: Admin failed to issue an invoice with non-existing invoice on the system
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And these invoices has "MANUAL" type
    And these invoice for students have default payment method "CONVENIENCE_STORE"
    And one invoice ID is added to the request but is non-existing
    When "school admin" bulk issue invoices using v2 endpoint with payment method "BULK_ISSUE_CONVENIENCE_STORE" and "TODAY" "TODAY+1"
    Then receives "InvalidArgument" status code

  Scenario Outline: Admin bulk issue invoices successfully and payment sequence number is manually set
    Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_PaymentSequenceNumberManualSetting"
    And there are "<existing-invoices>" preexisting number of existing invoices with "DRAFT" status
    And these invoices has "<invoice-type>" type
    And these invoice for students have default payment method "<default-payment-method>"
    When "<signed-in user>" bulk issue invoices using v2 endpoint with payment method "<bulk-issue-payment-method>" and "<due-date>" "<expiry-date>"
    Then receives "OK" status code
    And invoices status is updated to "ISSUED" status
    And payments exported tag is set to "false"
    And there are pending payment records for students created with payment method "<payment-method>" and "<due-date>" "<expiry-date>"
    And action log record for each invoice is recorded with "INVOICE_BULK_ISSUED" action log type
    And bulk payment record is created successfully with payment method "<bulk-payment-method>"

    Examples:
      | signed-in user | existing-invoices | bulk-issue-payment-method    | default-payment-method | payment-method    | due-date | expiry-date | invoice-type | bulk-payment-method            |
      | school admin   | 10                | BULK_ISSUE_DEFAULT_PAYMENT   | DIRECT_DEBIT           | DIRECT_DEBIT      | TODAY    | TODAY+1     | SCHEDULED    | BULK_PAYMENT_DEFAULT_PAYMENT   |
      | hq staff       | 20                | BULK_ISSUE_DEFAULT_PAYMENT   | CONVENIENCE_STORE      | CONVENIENCE_STORE | TODAY+1  | TODAY+2     | MANUAL       | BULK_PAYMENT_DEFAULT_PAYMENT   |
      | hq staff       | 10                | BULK_ISSUE_CONVENIENCE_STORE |                        | CONVENIENCE_STORE | TODAY+2  | TODAY+3     | MANUAL       | BULK_PAYMENT_CONVENIENCE_STORE |