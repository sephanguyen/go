@blocker
Feature: Kafka sync bill item

    Scenario: Kafka insert sync bill item
        When admin inserts a bill item record to fatima with status "<status>"
        Then invoicemgmt bill item table will be updated
    
    Examples:
        | status   |
        | BILLED   |
        | INVOICED |
        | PENDING  |

    Scenario: Kafka delete sync bill item
        Given there is an existing bill item on fatima with status "<status>"
        And this bill item is sync to invoicemgmt
        When admin deletes this bill item record on fatima
        Then this bill item on invoicemgmt will be deleted
    
    Examples:
        | status   |
        | BILLED   |
        | INVOICED |
        | PENDING  |