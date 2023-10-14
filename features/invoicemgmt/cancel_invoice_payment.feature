@blocker
Feature: Cancel Invoice
  Background:
    Given there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database

  Scenario: HQ Admin cancels cancels an invoice successfully
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And bill item has "<bill-item-previous-status>" previous status
    And there is "pending" payment history
    When admin cancels an invoice with "sample remark" remarks
    Then receives "OK" status code
    And invoice has "FAILED" invoice status
    And bill item has "invoiced" bill item status
    And latest payment record has "FAILED" payment status
    And action log record is recorded
    And action log has failed action

    Examples:
      | signed-in user      | bill-item-previous-status |
      | school admin        | pending                   |
      | hq staff            | billed                    |

  Scenario: HQ Admin cancels cancels an invoice successfully and empty remarks
    Given "<signed-in user>" logins to backoffice app
    And there is an existing invoice with "ISSUED" invoice status with bill item
    And bill item has "<bill-item-previous-status>" previous status
    And there is "pending" payment history
    When admin cancels an invoice with "" remarks
    Then receives "OK" status code
    And invoice has "FAILED" invoice status
    And bill item has "invoiced" bill item status
    And latest payment record has "FAILED" payment status
    And action log record is recorded
    And action log has failed action

    Examples:
      | signed-in user      | bill-item-previous-status |
      | school admin        | pending                   |
      | hq staff            | billed                    |

  Scenario: HQ Admin failed to cancel an invoice with invalid status
    Given "<signed-in user>" logins to backoffice app
    And there is a student that has bill item with status "INVOICED"
    And bill item exists in invoicemgmt database
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    When admin cancels an invoice with "sample remark" remarks
    Then receives "InvalidArgument" status code

    Examples:
        | signed-in user      | invoice-status |
        | school admin        | PAID           |
        | school admin        | VOID           |
        | school admin        | REFUNDED       |
        | school admin        | FAILED         |
        | school admin        | DRAFT          |
        | hq staff            | PAID           |
        | hq staff            | VOID           |
        | hq staff            | REFUNDED       |
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
