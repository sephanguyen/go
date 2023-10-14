@blocker
Feature: Sync discount table from fatima to invoicemgmt

  Scenario: Add record in discount in fatima and sync data to invoicemgmt
    When a discount record is inserted in fatima
    Then this discount record must be recorded in invoicemgmt