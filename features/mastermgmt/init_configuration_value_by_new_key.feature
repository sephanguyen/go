Feature: Init configuration value by new config_key

  Scenario Outline: Init configuration value by new config_key
    Given any org and config key in DB
    When a new "<type>" config key inserted in to DB
    Then new values of the new "<type>" config key are added for all existing org
    Examples:
      | type     |
      | internal |
      | external |