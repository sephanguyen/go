Feature: Creation of lesson chat
  As a lesson participants, i can see all live lesson chats available

  Background: A teacher and a student joined conversation
    Given resource path of school "Manabie" is applied
    Given a lesson conversation with "0" teachers and "2" students
    And a teacher joins lesson creating new lesson session

  Scenario: LiveLessonConversationDetail
    Then teacher sees correct info calling LiveLessonConversationDetail


  Scenario: old session with messages, start new session has seen status, empty latest message
    Given the first teacher sends "1" message with content "hello message" to live lesson chat
    And teacher sees correct latest message calling LiveLessonConversationDetail
    And a second teacher joins lesson without refreshing lesson session
    And the second teacher sees "unseen" status calling LiveLessonConversationDetail
    # refresh session to check again
    When a second teacher joins lesson refreshing lesson session
    Then the second teacher sees "seen" status calling LiveLessonConversationDetail
    And teacher sees empty latest message calling LiveLessonConversationDetail

  Scenario: new session without message has seen status, empty latest message
    Then teacher sees empty latest message calling LiveLessonConversationDetail
    And teacher sees "seen" status calling LiveLessonConversationDetail

  Scenario: Seening lesson conversation
    Given students join lesson without refreshing lesson session
    When a second teacher joins lesson without refreshing lesson session
    And the first teacher sends "1" message with content "hello message" to live lesson chat
    And the second teacher sees "unseen" status calling LiveLessonConversationDetail
    And students sees "unseen" status calling LiveLessonConversationDetail
    When The second teacher in lesson seen the conversation
    And students in lesson seen the conversation
    Then the second teacher sees "seen" status calling LiveLessonConversationDetail
    And students sees "seen" status calling LiveLessonConversationDetail

  @blocker
  Scenario Outline: student sends msg, student and teacher receives message
    Given a second teacher joins lesson without refreshing lesson session
    And students join lesson without refreshing lesson session
    When a student sends "1" message with content "<content>" to live lesson chat
    Then the "second teacher" in lesson receives "1" message with type "text" with content "<content>"
    And the "students" in lesson receives "1" message with type "text" with content "<content>"
    Examples:
      | content     |
      | hello world |
      | lorem ipsum |

  @blocker
  Scenario Outline: Teacher sends msg, student and teacher receives message
    Given students join lesson without refreshing lesson session
    And a second teacher joins lesson without refreshing lesson session
    When the first teacher sends "1" message with content "<content>" to live lesson chat
    Then the "second teacher" in lesson receives "1" message with type "text" with content "<content>"
    And the "students" in lesson receives "1" message with type "text" with content "<content>"
    Examples:
      | content     |
      | hello world |
      | lorem ipsum |

  # logic of whether to refresh lesson session or not is at client side.
  Scenario Outline: users do not see old session messages
    Given the first teacher sends "<num old message>" message with content "<content 1>" to live lesson chat
    And a second teacher joins lesson refreshing lesson session
    And students join lesson without refreshing lesson session
    When the first teacher sends "<num new message>" message with content "<content 2>" to live lesson chat
    Then the "second teacher" in lesson receives "<num new message>" message with type "text" with content "<content 2>"
    And the "students" in lesson receives "<num new message>" message with type "text" with content "<content 2>"
    And the second teacher sees "<num new message>" messages with content "<content 2>" calling LiveLessonConversationMessages
    Examples:
      | content 1   | content 2     | num old message | num new message |
      | hello world | hello world 2 | 3               | 4               |
      | lorem ipsum | lorem ipsum 2 | 5               | 6               |

  Scenario Outline: offline user receives silence notification
    Given a second teacher joins lesson but not subscribe stream
    And "students in lesson" device tokens exist in DB
    And "second teacher in lesson" device tokens exist in DB
    When the first teacher sends "1" message with content "<content>" to live lesson chat
    And the "students" in lesson receive silent notification with content "<content>"
    And the "second teacher" in lesson receive silent notification with content "<content>"
    Examples:
      | content     |
      | hello world |
      | lorem ipsum |
