@blocker
Feature: Payment Update Bill Items Status Endpoint Check
  Scenario: Update Bill Items Status from payment check successfully
      Given there is an existing bill items created on payment
      When payment endpoint is called to update these bill items status
      Then receives "OK" status code