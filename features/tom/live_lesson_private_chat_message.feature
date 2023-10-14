Feature: Create Live Lesson Private Chat

  Background: lesson conversation background
    Given resource path of school "Manabie" is applied
    And a lesson conversation with "2" teachers and "1" students

  Scenario Outline: user can see new messages from live lesson private conversation
    Given "<user 1>" create new live lesson private conversation with "<user 2>"
    When "<user 1>" sends "<num of message>" message to the live lesson private conversation with content "<content>"
    Then "<user 2>" sees "<num of message>" messages with content "<content>" when get live lesson private conversation messages

    Examples: 
      | user 1  | user 2  | content     | num of message |
      | student | teacher | hello world |              5 |
      | teacher | student | lorem ipsum |             10 |
      | teacher | teacher | hello world |             50 |

  Scenario Outline: users do not see old session messages in private conversations
    Given "<user 1>" create new live lesson private conversation with "<user 2>"
    And "<user 1>" sends "<num of old messages>" message to the live lesson private conversation with content "<old content>"
    When "<user 2>" refresh live lesson session for private conversation
    And "<user 1>" sends "<num of new messages>" message to the live lesson private conversation with content "<new content>"
    Then "<user 2>" sees "<num of new messages>" messages with content "<new content>" when get live lesson private conversation messages

    Examples: 
      | user 1  | user 2  | old content | new content     | num of old messages | num of new messages |
      | student | teacher | hello world | hello world new |                   5 |                  10 |
      | teacher | student | lorem ipsum | lorem ipsum new |                  10 |                  20 |
      | teacher | teacher | hello world | hello world new |                  15 |                  50 |

  Scenario Outline: users do not see old session messages in multiple private conversations
    Given multiple teacher create new live lesson private conversations with a student
    And multiple teacher sends "<num of old messages>" message to the live lesson private conversation with content "<old content>"
    When "<user>" refresh live lesson session for private conversation
    And multiple teacher sends "<num of new messages>" message to the live lesson private conversation with content "<new content>"
    Then "<user>" sees "<num of new messages>" messages with content "<new content>" when get messages in all private conversations

    Examples: 
      | user    | old content | new content     | num of old messages | num of new messages |
      | student | hello world | hello world new |                  15 |                  30 |
      | student | hello world | lorem ipsum new |                  30 |                  50 |

  Scenario: verify latest start time of all conversations is the same date
    Given multiple teacher create new live lesson private conversations with a student
    When "student" refresh live lesson session for private conversation
    Then live lesson conversation and all private conversations have latest start time with the same date
