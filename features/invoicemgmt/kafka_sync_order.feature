@blocker
Feature: Sync order table from fatima to invoicemgmt

  Scenario: Add record to order from fatima and sync to invoicemgmt
    When an order record is inserted into fatima
    Then this order record must be recorded in invoicemgmt