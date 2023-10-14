@quarantined
Feature: Issue Invoice without payment
  As an HQ manager or admin
  I am able to update the status of the status of the invoice to INVOICE_ISSUED and will not create a payment

  Background:
    Given there is an existing invoice
    And unleash feature flag is "disable" with feature name "Invoice_InvoiceManagement_BackOffice_SingleIssueInvoiceWithPayment"

  Scenario Outline: Admin issues an invoice successfully
    Given invoice has draft invoice status
    When "<signed-in user>" issues invoice using v2 endpoint
    Then receives "OK" status code
    And invoice status is updated to "ISSUED" status
    And invoice exported tag is set to "false"
    And this invoice has "no existing" payment with "PENDING" status
    And action log record is recorded with "INVOICE_ISSUED" action log type

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |

  Scenario Outline: Admin issues an invoice with zero total amount and invoice status updated to PAID
    Given invoice has draft invoice status
    And this invoice has "0" total amount
    When "<signed-in user>" issues invoice using v2 endpoint
    Then receives "OK" status code
    And invoice status is updated to "PAID" status
    And invoice exported tag is set to "false"
    And this invoice has "no existing" payment with "PENDING" status
    And action log record is recorded with "INVOICE_ISSUED" action log type

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |

  Scenario Outline: Admin failed to issue an invoice with non-existing invoice ID
    Given invoice ID is non-existing
    When "<signed-in user>" issues invoice using v2 endpoint
    Then receives "Internal" status code
    And no payment history is recorded

    Examples:
      | signed-in user |
      | school admin   |
      | hq staff       |