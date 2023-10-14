@quarantined
Feature: Order data sync to Elasticsearch by Kafka
    Scenario Outline: Sync order record to order index in Elasticsearch
        Given prepare order data for elastic sync
        When a record is "<operation>" in order table
        Then the record "<operation>" must be reflected in ES order index

    Examples:
        | operation |
        | update    |
        | delete    |
