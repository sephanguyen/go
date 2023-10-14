
Feature: Void Invoice with payment amount set to zero
  As an HQ manager or admin
  I am able to void an invoice

  Scenario Outline: Admin voids an invoice successfully
    Given "<signed-in user>" logins to backoffice app
    And there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    And bill item has "<bill-item-previous-status>" previous status
    And has billing date "<billing-date-compared-to-today>" today
    And there is "<payment-history>" payment history
    When admin voids an invoice with "<remarks>" remarks using v2 endpoint
    Then receives "OK" status code
    And invoice has "VOID" invoice status
    And bill item has "<bill-item-new-status>" bill item status
    And latest payment record has "<payment-status>" payment status and amount zero
    And action log record is recorded

    Examples:
      | signed-in user | bill-item-previous-status | invoice-status | payment-history | remarks | payment-status | billing-date-compared-to-today | bill-item-new-status |
      | school admin   | billed                    | DRAFT          | none            | any     | none           | after                          | billed               |
      | school admin   | pending                   | ISSUED         | PENDING         | none    | FAILED         | after                          | pending              |
      | hq staff       | pending                   | ISSUED         | FAILED          | none    | FAILED         | before                         | billed               |
      | hq staff       | pending                   | DRAFT          | PENDING         | none    | FAILED         | same                           | billed               |

  Scenario Outline: Admin failed to void an invoice with invalid status
    Given "<signed-in user>" logins to backoffice app
    And there is a student that has bill item with status "PENDING"
    And bill item exists in invoicemgmt database
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    When admin voids an invoice with "any" remarks
    Then receives "InvalidArgument" status code

    Examples:
      | signed-in user | invoice-status | remarks |
      | school admin   | PAID           | none    |
      | school admin   | VOID           | any     |
      | hq staff       | REFUNDED       | any     |

  Scenario Outline: Admin voids an invoice successfully and bulk payment was updated to CANCELLED
    Given "school admin" logins to backoffice app
    And there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And bill item has "pending" previous status
    And has billing date "before" today
    And there is "PENDING" payment history
    And belongs in bulk with "3" other "FAILED" payments with "CONVENIENCE_STORE" payment method
    When admin voids an invoice with "any" remarks using v2 endpoint
    Then receives "OK" status code
    And invoice has "VOID" invoice status
    And bill item has "billed" bill item status
    And latest payment record has "FAILED" payment status and amount zero
    And action log record is recorded
    And bulk payment record has "BULK_PAYMENT_CANCELLED" status