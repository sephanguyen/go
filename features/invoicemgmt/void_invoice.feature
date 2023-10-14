
@quarantined
Feature: Void Invoice
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
    When admin voids an invoice with "<remarks>" remarks
    Then receives "OK" status code
    And invoice has "VOID" invoice status
    And bill item has "<bill-item-new-status>" bill item status
    And latest payment record has "<payment-status>" payment status
    And action log record is recorded

    Examples:
        | signed-in user      | bill-item-previous-status | invoice-status | payment-history        | remarks | payment-status | billing-date-compared-to-today | bill-item-new-status |
        | school admin        | billed                    | DRAFT          | none                   | any     | none           | after                          | billed               |
        | school admin        | billed                    | ISSUED         | FAILED-PENDING         | any     | FAILED         | after                          | billed               |
        | school admin        | billed                    | FAILED         | FAILED                 | any     | FAILED         | same                           | billed               |
        | school admin        | pending                   | DRAFT          | none                   | none    | none           | same                           | billed               |
        | school admin        | pending                   | ISSUED         | PENDING                | none    | FAILED         | after                          | pending              |
        | school admin        | pending                   | FAILED         | FAILED-FAILED          | none    | FAILED         | before                         | billed               |
        | school admin        | billed                    | ISSUED         | FAILED                 | none    | FAILED         | same                           | billed               |
        | school admin        | billed                    | ISSUED         | FAILED-FAILED-PENDING  | any     | FAILED         | before                         | billed               |
        | hq staff            | billed                    | DRAFT          | none                   | any     | none           | before                         | billed               |
        | hq staff            | billed                    | ISSUED         | FAILED-PENDING         | any     | FAILED         | before                         | billed               |
        | hq staff            | billed                    | FAILED         | FAILED                 | any     | FAILED         | before                         | billed               |
        | hq staff            | pending                   | DRAFT          | none                   | none    | none           | before                         | billed               |
        | hq staff            | pending                   | ISSUED         | PENDING                | none    | FAILED         | same                           | billed               |
        | hq staff            | pending                   | FAILED         | FAILED-FAILED          | none    | FAILED         | after                          | pending              |
        | hq staff            | billed                    | ISSUED         | FAILED                 | none    | FAILED         | after                          | billed               |
        | hq staff            | billed                    | ISSUED         | FAILED-FAILED-PENDING  | any     | FAILED         | after                          | billed               |

  Scenario Outline: Admin failed to void an invoice with invalid status
    Given "<signed-in user>" logins to backoffice app
    And there is a student that has bill item with status "PENDING"
    And bill item exists in invoicemgmt database
    And there is an existing invoice with "<invoice-status>" invoice status with bill item
    When admin voids an invoice with "any" remarks
    Then receives "InvalidArgument" status code

    Examples:
        | signed-in user      | invoice-status |
        | school admin        | PAID           |
        | school admin        | VOID           |
        | school admin        | REFUNDED       |
        | hq staff            | PAID           |
        | hq staff            | VOID           |
        | hq staff            | REFUNDED       |

  Scenario Outline: Admin failed to void a non-existing invoice
    Given "<signed-in user>" logins to backoffice app
    Given invoice ID is non-existing
    When admin voids an invoice with "any" remarks
    Then receives "Internal" status code

    Examples:
        | signed-in user      |
        | school admin        |
        | hq staff            |

  Scenario Outline: Unauthorized users failed to void an invoice
    Given "<unauthorized-user>" logins to backoffice app
    When admin voids an invoice with "any" remarks
    Then receives "<status-code>" status code

    Examples:
      | unauthorized-user | status-code      |
      | student           | PermissionDenied |
      | parent            | PermissionDenied |
      | teacher           | PermissionDenied |
      | unauthenticated   | Unauthenticated  |
