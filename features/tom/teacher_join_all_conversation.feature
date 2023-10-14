Feature: Teacher join all conversation

  Background: default manabie resource path
    Given resource path of school "Manabie" is applied

  Scenario Outline: teacher join all conversation
    And a valid "<user>" token
    Given random new conversations created
    When "<user>" joins all conversations
    Then returns "OK" status code
    And "<user>" must be member of all conversations with specific schools
    And system must send "joined" conversation message

    Examples:
      | user         |
      | teacher      |
      | school admin |

  Scenario: the teacher has joined some conversation before and choose join all conversation
    And a valid "teacher" token
    Given random new conversations created
    And the teacher joins some conversations
    When teacher joins all conversations
    Then returns "OK" status code
    And teacher must be member of all conversations with specific schools
    And system must send only "joined" message which unjoined conversations before

  Scenario: Teacher does not join conversation with locations not in request
    Given a new school is created with location "default"
    And a valid "teacher" token
    And locations "loc 1,loc 2" children of location "default" are created
    Given "2" new conversations created for each locations "loc 1,loc 2"
    When teacher joins all conversations in locations "loc 1"
    Then teacher "must" be member of "2" conversations in locations "loc 1"
    And teacher "must not" be member of "2" conversations in locations "loc 2"

  Scenario: Joining all multiple locations
    Given a new school is created with location "default"
    And a valid "teacher" token
    And locations "loc 1,loc 2,loc 3" children of location "default" are created
    Given "1" new conversations created for each locations "loc 1,loc 2,loc 3"
    When teacher joins all conversations in locations "loc 1,loc 3"
    Then teacher "must" be member of "2" conversations in locations "loc 1,loc 3"
    Then teacher "must not" be member of "1" conversations in locations "loc 2"

  Scenario: Joining with parent location also join all children locations
    Given a new school is created with location "default"
    And a valid "teacher" token
    And locations "loc 1,loc 2" children of location "default" are created
    And locations "loc 3" children of location "loc 2" are created
    Given "1" new conversations created for each locations "loc 1,loc 2,loc 3"
    When teacher joins all conversations in locations "loc 2"
    Then teacher "must" be member of "2" conversations in locations "loc 2,loc 3"
    Then teacher "must not" be member of "1" conversations in locations "loc 1"
