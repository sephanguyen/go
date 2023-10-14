Feature: Convert media

  Scenario Outline: Teacher convert media
    Given a list of media
    And "<signed-in user>" signin system
    When user converts media to image
    Then media conversion tasks must be created

    Examples:
      | signed-in user |
      | teacher        |
      | school admin   |

  Scenario: Handle media conversion tasks events
    Given a list of media conversion tasks
    When our system receives a finished conversion event
    Then finished conversion tasks must be updated
