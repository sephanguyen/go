@major
Feature: Refund Invoice
    As an HQ manager or admin
    I am able to refund an invoice

    Background:
        Given there is a student that has bill item with status "INVOICED"
        And bill item exists in invoicemgmt database

    Scenario Outline: Admin refund an invoice successfully
        Given there is an existing invoice with "ISSUED" invoice status with bill item
        And this invoice has "<invoice-total>" total amount
        And "<signed-in user>" logins to backoffice app
        And admin refunds an invoice
        And sets refund method "<refund-method>" in refund invoice request
        And sets amount same with invoice outstanding balance in refund invoice request
        When admin submits the refund invoice form with remarks "<remarks>"
        Then receives "OK" status code
        And invoice status is updated to "REFUNDED" status
        And invoice outstanding_balance set to "0.00"
        And invoice amount_refunded set to "<invoice-total>"
        And action log record is recorded with "INVOICE_REFUNDED" action log type

        Examples:
            | signed-in user | refund-method | invoice-total | remarks        |
            | school admin   | CASH          | -100          | sample-remarks |
            | school admin   | BANK_TRANSFER | -50           | sample-remarks |
            | hq staff       | CASH          | -1000         |                |
            | hq staff       | BANK_TRANSFER | -5000         |                |

    Scenario Outline: Admin refund an invoice with zero or positive total amount
        Given there is an existing invoice with "ISSUED" invoice status with bill item
        And this invoice has "<invoice-total>" total amount
        And "<signed-in user>" logins to backoffice app
        And admin refunds an invoice
        And sets refund method "<refund-method>" in refund invoice request
        And sets amount same with invoice outstanding balance in refund invoice request
        When admin submits the refund invoice form with remarks "<remarks>"
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | refund-method | invoice-total | remarks        |
            | school admin   | CASH          | 100           | sample-remarks |
            | school admin   | BANK_TRANSFER | 50            | sample-remarks |
            | hq staff       | CASH          | 0             |                |
            | hq staff       | BANK_TRANSFER | 0             |                |

    Scenario Outline: Admin refund an invoice when the given amount is not equal to invoice total
        Given there is an existing invoice with "ISSUED" invoice status with bill item
        And this invoice has "<invoice-total>" total amount
        And "<signed-in user>" logins to backoffice app
        And admin refunds an invoice
        And sets refund method "<refund-method>" in refund invoice request
        And sets amount to "<amount>" in refund invoice request
        When admin submits the refund invoice form with remarks "<remarks>"
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | refund-method | invoice-total | amount | remarks        |
            | school admin   | CASH          | -100          | -50    | sample-remarks |
            | hq staff       | BANK_TRANSFER | -100          | -50    | sample-remarks |

    Scenario Outline: Admin refund an invoice with invalid invoice status
        Given there is an existing invoice with "<invoice-status>" invoice status with bill item
        And "<signed-in user>" logins to backoffice app
        And admin refunds an invoice
        And sets refund method "<refund-method>" in refund invoice request
        And sets amount same with invoice outstanding balance in refund invoice request
        When admin submits the refund invoice form with remarks "<remarks>"
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | refund-method | invoice-status | remarks        |
            | school admin   | CASH          | PAID           | sample-remarks |
            | school admin   | BANK_TRANSFER | VOID           | sample-remarks |
            | school admin   | CASH          | REFUNDED       | sample-remarks |
            | hq staff       | BANK_TRANSFER | FAILED         | sample-remarks |
            | hq staff       | CASH          | DRAFT          | sample-remarks |
