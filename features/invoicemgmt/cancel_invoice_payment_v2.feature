@major
Feature: Cancel Invoice Payment V2
  Background:
    Given there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database

  Scenario: HQ Admin cancels cancels an invoice successfully
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And there is "PENDING" payment history with "<payment-method>" payment method
    And this payment has exported status "<exported-status>"
    When admin cancels an invoice with "<remarks>" remarks using v2 endpoint
    Then receives "OK" status code
    And invoice remains "ISSUED" invoice status
    And latest payment record has "FAILED" payment status and amount zero
    And action log record is recorded with "PAYMENT_CANCELLED" action and "<remarks>" remarks

    Examples:
      | signed-in user      | remarks       | payment-method    | exported-status |
      | school admin        | sample remark | CASH              | TRUE            |
      | school admin        |               | CASH              | FALSE           |
      | hq staff            |               | BANK_TRANSFER     | TRUE            |
      | hq staff            | sample remark | BANK_TRANSFER     | FALSE           |
      | hq staff            | test remark   | CONVENIENCE_STORE | TRUE            |
      | hq staff            | test remark   | CONVENIENCE_STORE | FALSE           |
      | hq staff            | test remark   | DIRECT_DEBIT      | FALSE           |

  Scenario: HQ Admin cancels an invoice with payment successfully belong in bulk payment updated to cancelled
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And there is "PENDING" payment history with "<payment-method>" payment method
    And this payment has exported status "FALSE"
    And belongs in bulk with "<other-payment-count>" other "FAILED" payments with "<payment-method>" payment method
    When admin cancels an invoice with "<remarks>" remarks using v2 endpoint
    Then receives "OK" status code
    And invoice remains "ISSUED" invoice status
    And latest payment record has "FAILED" payment status and amount zero
    And action log record is recorded with "PAYMENT_CANCELLED" action and "<remarks>" remarks
    And bulk payment record has "BULK_PAYMENT_CANCELLED" status

    Examples:
      | signed-in user      | remarks       | payment-method    | other-payment-count |
      | school admin        | sample remark | CASH              | 2                   |
      | hq staff            | sample remark | BANK_TRANSFER     | 1                   |
      | hq staff            | test remark   | CONVENIENCE_STORE | 1                   |
      | school admin        | test remark   | DIRECT_DEBIT      | 2                   |
  
  Scenario: HQ Admin cancels an invoice with payment successfully belong in bulk payment with no changes
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And there is "PENDING" payment history with "<payment-method>" payment method
    And this payment has exported status "<exported-status>"
    And belongs in bulk with "<other-payment-count>" other "PENDING" payments with "<payment-method>" payment method
    When admin cancels an invoice with "<remarks>" remarks using v2 endpoint
    Then receives "OK" status code
    And invoice remains "ISSUED" invoice status
    And latest payment record has "FAILED" payment status and amount zero
    And action log record is recorded with "PAYMENT_CANCELLED" action and "<remarks>" remarks
    And bulk payment record has "BULK_PAYMENT_PENDING" status

    Examples:
      | signed-in user      | remarks       | payment-method    | other-payment-count |
      | school admin        | sample remark | CASH              | 2                   |
      | hq staff            | sample remark | BANK_TRANSFER     | 1                   |
      | hq staff            | test remark   | CONVENIENCE_STORE | 1                   |
      | school admin        | test remark   | DIRECT_DEBIT      | 2                   |

  Scenario: HQ Admin failed to cancel an invoice with already exported direct debit payment 
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And there is "PENDING" payment history with "DIRECT_DEBIT" payment method
    And this payment has exported status "TRUE"
    When admin cancels an invoice with "sample remark" remarks using v2 endpoint
    Then receives "Internal" status code

    Examples:
      | signed-in user      |
      | school admin        |

  Scenario: HQ Admin failed to cancel an invoice with invalid status
    Given "<signed-in user>" logins to backoffice app
    And there is a student that has bill item with status "INVOICED"
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    When admin cancels an invoice with "sample remark" remarks using v2 endpoint
    Then receives "InvalidArgument" status code

    Examples:
        | signed-in user      | invoice-status |
        | school admin        | PAID           |
        | school admin        | VOID           |
        | school admin        | REFUNDED       |
        | hq staff            | FAILED         |
        | hq staff            | DRAFT          |

  Scenario: HQ Admin failed to cancel a non-existing invoice
    Given "<signed-in user>" logins to backoffice app
    And invoice ID is non-existing
    When admin cancels an invoice with "sample remark" remarks
    Then receives "Internal" status code

    Examples:
      | signed-in user      |
      | school admin        |
      | hq staff            |
