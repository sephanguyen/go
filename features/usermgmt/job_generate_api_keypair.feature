Feature: Run job to generate API Key

  Scenario Outline: Run job to generate API Key
    Given a signed in "school admin"
    When system run job to generate API Key with organization "<organization>"
    Then API keypair is created successfully

    Examples:
      | organization   |
      | MANABIE_SCHOOL |
