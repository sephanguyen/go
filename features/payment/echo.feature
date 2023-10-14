Feature: Echo a message

  Scenario: Echo a message
    Given an echo message
    When "<signed-in user>" echo a message
    Then the message is echoed
    And receives "OK" status code

    Examples:
      | signed-in user |
      | school admin   |
