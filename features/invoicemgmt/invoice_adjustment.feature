Feature: Invoice Adjustment
  As an HQ manager or admin
  I can adjust the total and subtotal amount of an invoice that has draft status

  @blocker
  Scenario Outline: Admin creates invoice adjustment record successfully
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And this invoice has "<invoice-total>" total amount
    When "<signed-in user>" logins to backoffice app
    And adds "<record-count>" invoice adjustment with "<adjust-amount-to-create>" amount
    And apply the adjustment on the invoice
    Then receives "OK" status code
    And invoice total subtotal and outstanding balance are correctly updated to "<expected-amount>" amount
    And action log record is recorded with "INVOICE_ADJUSTED" action log type

    Examples:
      | signed-in user | invoice-total | record-count | adjust-amount-to-create | expected-amount |
      | school admin   | 3000          | 1            | 550                     | 3550            |
      | hq staff       | 1000          | 2            | 1000&700                | 2700            |
      | hq staff       | 1000          | 1            | -500                    | 500             |

  @blocker
  Scenario Outline: Admin edits invoice adjustment record successfully
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And this invoice has "<invoice-total>" total amount
    And there are "<record-count>" created invoice adjustment with "<adjust-amount-to-create>" amount
    When "<signed-in user>" logins to backoffice app
    And edits "<edit-record-count>" existing invoice adjustment with "<existing-amount>" amount updated to "<updated-amount>" amount
    And apply the adjustment on the invoice
    Then receives "OK" status code
    And invoice total subtotal and outstanding balance are correctly updated to "<expected-amount>" amount
    And action log record is recorded with "INVOICE_ADJUSTED" action log type

    Examples:
      | signed-in user | invoice-total | record-count | edit-record-count | adjust-amount-to-create | expected-amount | existing-amount | updated-amount |
      | school admin   | 200           | 1            | 1                 | 100                     | 400             | 100             | 200            |
      | hq staff       | 1500          | 2            | 1                 | 570&300                 | 2470            | 300             | 400            |
      | hq staff       | 50            | 2            | 2                 | 5&45                    | 115             | 5&45            | 10&55          |
      | school admin   | -10           | 1            | 1                 | 200                     | 290             | 200             | 300            |
      | school admin   | 100           | 1            | 1                 | 100                     | -100            | 100             | -200           |

  @blocker
  Scenario Outline: Admin deletes invoice adjustment record successfully
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And this invoice has "<invoice-total>" total amount
    And there are "<record-count>" created invoice adjustment with "<adjust-amount-to-create>" amount
    When "<signed-in user>" logins to backoffice app
    And deletes "<delete-record-count>" existing invoice adjustment with "<existing-amount>" amount
    And apply the adjustment on the invoice
    Then receives "OK" status code
    And invoice total subtotal and outstanding balance are correctly updated to "<expected-amount>" amount
    And action log record is recorded with "INVOICE_ADJUSTED" action log type

    Examples:
      | signed-in user | invoice-total | record-count | delete-record-count | adjust-amount-to-create | expected-amount | existing-amount |
      | school admin   | 100           | 1            | 1                   | 100                     | 100             | 100             |
      | hq staff       | 150           | 2            | 1                   | 15&25                   | 165             | 25              |
      | hq staff       | 80            | 2            | 2                   | 20&80                   | 80              | 20&80           |
      | school admin   | -5            | 1            | 1                   | 40                      | -5              | 40              |

  @blocker
  Scenario Outline: Admin creates, edits and deletes invoice adjustment record successfully
    Given there are "1" preexisting number of existing invoices with "DRAFT" status
    And this invoice has "150" total amount
    And there are "2" created invoice adjustment with "30&60" amount
    When "school admin" logins to backoffice app
    And edits "1" existing invoice adjustment with "60" amount updated to "90" amount
    And adds "1" invoice adjustment with "60" amount
    And deletes "1" existing invoice adjustment with "30" amount
    And apply the adjustment on the invoice
    Then receives "OK" status code
    And invoice total subtotal and outstanding balance are correctly updated to "300" amount
    And action log record is recorded with "INVOICE_ADJUSTED" action log type

  @major
  Scenario Outline: Admin requests invoice adjustment record with invalid invoice status
    Given there are "1" preexisting number of existing invoices with "<invoice-status>" status
    And this invoice has "<invoice-total>" total amount
    When "<signed-in user>" logins to backoffice app
    And adds "1" invoice adjustment with "<adjust-amount-to-create>" amount
    And apply the adjustment on the invoice
    Then receives "FailedPrecondition" status code

    Examples:
      | signed-in user | invoice-total | record-count | adjust-amount-to-create | invoice-status |
      | school admin   | 3             | 1            | 2                       | ISSUED         |
      | hq staff       | 3             | 1            | 2                       | PAID           |
      | hq staff       | 3             | 1            | 2                       | FAILED         |
      | hq staff       | 3             | 1            | 2                       | VOID           |
