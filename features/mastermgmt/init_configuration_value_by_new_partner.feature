Feature: Init configuration value by new organization

  Scenario Outline: Init configuration value by new organization
    Given any org and config key in DB
    When a new org inserted in to DB
    Then new values of all existing config key are added for the new org