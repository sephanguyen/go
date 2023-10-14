@blocker
Feature: Manually Generate Invoice

    Scenario: HQ manager successfully creates multiple invoices
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<student_count>" students that has "1" bill item with status "<status>" and type "<bill_type>"
        And "no" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there are "<student_count>" student draft invoices created successfully
        And invoice bill item is created
        And there are no errors in response
        And invoice data is present in the response with count "<student_count>"
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | status  | student_count | bill_type                       |
            | school admin   | BILLED  | 3             | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | PENDING | 3             | BILLING_TYPE_ADJUSTMENT_BILLING |

    Scenario: HQ manager successfully creates multiple invoices that have multiple bill items
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<student_count>" students that has "<bill_item_count>" bill item with status "<status>" and type "<bill_type>"
        And "no" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there are "<student_count>" student draft invoices created successfully
        And invoice bill item is created
        And there are no errors in response
        And invoice data is present in the response with count "<student_count>"
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | status  | student_count | bill_item_count | bill_type                       |
            | school admin   | BILLED  | 3             | 3               | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | PENDING | 3             | 3               | BILLING_TYPE_ADJUSTMENT_BILLING |

    Scenario: HQ manager successfully creates invoices for valid bill items and return a response error for invalid one
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<valid_count>" students that has "1" bill item with status "<valid_status>" and type "<bill_type>"
        And there are "<invalid_count>" students that has "1" bill item with status "<invalid_status>" and type "<bill_type>"
        And "no" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there are "<valid_count>" student draft invoices created successfully
        And invoice bill item is created
        And there are "<invalid_count>" response error
        And invoice data is present in the response with count "<valid_count>"
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | valid_count | invalid_count | valid_status | invalid_status | bill_type                       |
            | school admin   | 3           | 3             | BILLED       | INVOICED       | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | 3           | 3             | PENDING      | CANCELLED      | BILLING_TYPE_ADJUSTMENT_BILLING |

    Scenario: HQ manager request with an empty invoices and receive InvalidArgument status code
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "0" students that has "0" bill item with status "BILLED" and type "<bill_type>"
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "InvalidArgument" status code

        Examples:
            | signed-in user | bill_type                       |
            | school admin   | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | BILLING_TYPE_ADJUSTMENT_BILLING |

    Scenario: HQ manager failed to create invoice due to bill item tagged with Review Required
        Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<student_count>" students that has "<bill_item_count>" bill item with status "<status>" and type "<bill_type>"
        And "one" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there is an error and no invoice in the response

        Examples:
            | signed-in user | status | student_count | bill_item_count | bill_type                    |
            | school admin   | BILLED | 1             | 3               | BILLING_TYPE_BILLED_AT_ORDER |


    # ----------------------------------- Payment_OrderManagement_BackOffice_Reviewed_Flag disabled --------------------------------

    @quarantined
    Scenario: HQ manager successfully creates multiple invoices even if bill items have review required tag
        Given unleash feature flag is "disable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<student_count>" students that has "1" bill item with status "<status>" and type "<bill_type>"
        And "all" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there are "<student_count>" student draft invoices created successfully
        And invoice bill item is created
        And there are no errors in response
        And invoice data is present in the response with count "<student_count>"
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | status  | student_count | bill_type                       |
            | school admin   | BILLED  | 3             | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | PENDING | 3             | BILLING_TYPE_ADJUSTMENT_BILLING |

    @quarantined
    Scenario: HQ manager successfully creates multiple invoices that have multiple bill items even if bill items have review required tag
        Given unleash feature flag is "disable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
        And "<signed-in user>" logins to backoffice app
        And there are "<student_count>" students that has "<bill_item_count>" bill item with status "<status>" and type "<bill_type>"
        And "all" bill item has review required tag
        When generateInvoice endpoint is called to create multiple invoice
        Then receives "OK" status code
        And there are "<student_count>" student draft invoices created successfully
        And invoice bill item is created
        And there are no errors in response
        And invoice data is present in the response with count "<student_count>"
        And each invoice has correct total amount and outstanding balance

        Examples:
            | signed-in user | status  | student_count | bill_item_count | bill_type                       |
            | school admin   | BILLED  | 3             | 3               | BILLING_TYPE_BILLED_AT_ORDER    |
            | hq staff       | PENDING | 3             | 3               | BILLING_TYPE_ADJUSTMENT_BILLING |
