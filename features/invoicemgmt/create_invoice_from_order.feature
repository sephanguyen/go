Feature: Create Invoice From Order

    Background:
        Given there are "5" existing students

    Scenario Outline: HQ manager successfully creates multiple invoices from list of order
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "<order-count>" orders with status "SUBMITTED" and "no-existing" review required tag
        And each of these orders have "<bill-item-count>" bill items with status "BILLED"
        And "all" billing items of order have billing date at day "-10"
        And there is invoice date scheduled at day "10"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "OK" status code
        And there are "5" student draft invoices created successfully
        And each invoice have "<bill-item-count>" bill items
        And these bill items have "INVOICED" billing status
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | order-count | bill-item-count |
            | school admin   | 1           | 3               |
            | hq staff       | 1           | 3               |

    Scenario Outline: HQ manager successfully creates multiple invoices from list of order and didn't include the invalid bill item
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "<order-count>" orders with status "SUBMITTED" and "no-existing" review required tag
        And each of these orders have "<bill-item-count>" bill items with status "PENDING"
        And "one" billing items of order have billing date at day "20"
        And there is invoice date scheduled at day "10"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "OK" status code
        And there are "5" student draft invoices created successfully
        And each invoice have "<total-bill-item-count>" bill items
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | order-count | bill-item-count | total-bill-item-count |
            | school admin   | 1           | 3               | 2                     |
            | hq staff       | 1           | 3               | 2                     |

    Scenario Outline: HQ manager successfully creates multiple invoices of students' multiple orders
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "<order-count>" orders with status "SUBMITTED" and "no-existing" review required tag
        And each of these orders have "<bill-item-count>" bill items with status "BILLED"
        And "all" billing items of order have billing date at day "-10"
        And there is invoice date scheduled at day "10"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "OK" status code
        And there are "5" student draft invoices created successfully
        And each invoice have "<total-bill-item-count>" bill items
        And these bill items have "INVOICED" billing status
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | order-count | bill-item-count | total-bill-item-count |
            | school admin   | 2           | 3               | 6                     |
            | hq staff       | 2           | 3               | 6                     |

    Scenario Outline: HQ manager successfully creates multiple invoices from list of order with adjument billing type
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "<order-count>" orders with status "SUBMITTED" and "no-existing" review required tag
        And each of these orders have "<bill-item-count>" bill items with status "BILLED"
        And "all" billing items of order have billing date at day "-10"
        And "<all-or-one>" billing items of order have adjustment price "<adjustment-price>"
        And there is invoice date scheduled at day "10"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "OK" status code
        And there are "5" student draft invoices created successfully
        And each invoice have "<bill-item-count>" bill items
        And these bill items have "INVOICED" billing status
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | order-count | bill-item-count | all-or-one | adjustment-price |
            | school admin   | 1           | 3               | all        | 8                |
            | hq staff       | 1           | 3               | one        | -8               |

    Scenario Outline: HQ manager successfully creates multiple invoices with invalid order status and review required tag
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "1" orders with status "<order-status>" and "<review-required>" review required tag
        And each of these orders have "1" bill items with status "BILLED"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | order-status | review-required |
            | school admin   | PENDING      | no-existing     |
            | hq staff       | VOIDED       | no-existing     |
            | school admin   | REJECTED     | no-existing     |
            | hq staff       | INVOICED     | no-existing     |
            | school admin   | SUBMITTED    | existing        |


    # ----------------------------------- Payment_OrderManagement_BackOffice_Reviewed_Flag disabled --------------------------------

    @quarantined
    Scenario Outline: HQ manager successfully creates multiple invoices from list of order even if bill items have review required tag
        Given unleash feature flag is "disable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And each of these students have "<order-count>" orders with status "SUBMITTED" and "existing" review required tag
        And each of these orders have "<bill-item-count>" bill items with status "BILLED"
        And "all" billing items of order have billing date at day "-10"
        And there is invoice date scheduled at day "10"
        And "<signed-in user>" logins to backoffice app
        When admin selects the order list
        And submits the create invoice from order request
        Then receives "OK" status code
        And there are "5" student draft invoices created successfully
        And each invoice have "<bill-item-count>" bill items
        And these bill items have "INVOICED" billing status
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | order-count | bill-item-count |
            | school admin   | 1           | 3               |
            | hq staff       | 1           | 3               |