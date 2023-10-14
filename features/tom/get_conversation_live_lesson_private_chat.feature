Feature: Get Conversation With Live Lesson Private Chat

  Background: lesson conversation background
    Given resource path of school "Manabie" is applied
    And a lesson conversation with "2" teachers and "1" students

  Scenario Outline: user can see conversation detail from live lesson private conversation 
    Given "<user 1>" create new live lesson private conversation with "<user 2>"
    And "<user 1>" sends "<num of message>" message to the live lesson private conversation with content "<content>"
    When "<user 2>" get the lesson private conversation detail 
    Then "<user 2>" sees the lesson private conversation returned with the correct data
    Examples: 
      | user 1  | user 2  | content     | num of message |
      | student | teacher | hello world |              5 |
      | teacher | student | lorem ipsum |             10 |
      | teacher | teacher | hello world |             50 |
