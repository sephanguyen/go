Feature: Access control for parents to some APIs

  Scenario Outline: A parent is allowed to access some APIs
    Given "parent" signin system
    When user calls "<api>" API
    Then returns "OK" status code

    Examples:
        | api                      |
        # | RetrieveLearningProgress |
        # | RetrieveStat             |
        | CountTotalLOsFinished    |