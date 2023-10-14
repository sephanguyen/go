Feature: Students/Parents chat group teachers

  Background: default manabie resource path
    Given resource path of school "Manabie" is applied

  Scenario: Student can view chat list
    Given a chat between a student and "1" teachers
    When student go to messages on learner app
    Then the screen displays 1 student chat

    # Done
  Scenario Outline: Student sends chat items when all participants are present
    Given a chat between a student and "<num teacher>" teachers
    And student and teachers are present
    When student sends "<message item>" item with content "<content>"
    Then teachers receive sent message
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |

    #Pending email notification
  Scenario Outline: Student sends chat items when all teachers are not present
    Given a chat between a student and "<num teacher>" teachers
    And student is present
    But teachers are not present
    When student sends "<message item>" item with content "<content>"
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |


  Scenario Outline: Student reading message marks message as read
    Given a chat between a student and "<num teacher>" teachers
    And student and teachers are present
    When a teacher sends "<message item>" item with content "<content>"
    And student seen the message
    Then teachers see the message has been read
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |

    #Done
  Scenario Outline: Teacher sends chat items to student when all participants are present
    Given a chat between a student and "<num teacher>" teachers
    And student and teachers are present
    When a teacher sends "<message item>" item with content "<content>"
    Then student receives sent message
    And other teachers receive sent message
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |

    #Pending email notification
  @wip
  Scenario: Student sends chat items when teachers have not join chat
    Given a chat of a student only
    And "<num teachers>" have not joined the chat
    When student sends "<message item>" item with content "<content>"
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |

    #Done
  Scenario Outline: Teacher sends chat items to student when students are not present
    Given a chat between a student and "<num teacher>" teachers
    And a "student" device token is existed in DB
    And teachers are present
    But student is not present
    When a teacher sends "<message item>" item with content "<content>"
    And student receives notification
    Examples:
      | message item | content        | num teacher |
      | text         | Hello world    | 1           |
      | file         | test-doc.pdf   | 2           |
      | image        | test-image.jpg | 3           |

    #Done
  Scenario: Parents with multiple kids have multiple chat group for each
    Given "2" student-teacher chats
    And account for parent of these kids is created
    Then chats are created for parent
    And each parent chat has name assigned to the kids' name

    #Done
  Scenario: Parents account exist and student is then created later, parents are added into the same chat group
    Given "2" parents account exist before "2" student accounts are created
    Then all parents are added into chat groups


  Scenario: Parent that shares the same students with previous parents are added into the same chat group
    Given chats between a parent and "1" teachers to manage "1" kids
    And another parent account is created with event "ParentAssignedToStudent"
    Then this parent is added in these chats

  Scenario: Parent can view chat list
    Given "3" student-teacher chats
    And account for parent of these kids is created
    Then chats are created for parent
    And parent can view "3" chats on learner app


  Scenario Outline: Parent can send chat items when teachers are not present
    Given a chat between "<num parent>" parents and "<num teacher>" teachers
    And parents are present
    But teachers are not present
    When a parent sends "<message item>" item with content "<content>"
    Then other parents receive sent message
    Examples:
      | num parent | message item | content        | num teacher |
      | 1          | text         | Hello world    | 1           |
      | 1          | file         | test-doc.pdf   | 2           |
      | 1          | image        | test-image.jpg | 2           |

    #Pending for email notification
  Scenario Outline: Teacher can send chat items to parent when all participants are present
    Given a chat between "<num parent>" parents and "<num teacher>" teachers
    And parents and teachers are present
    When a teacher sends "<message item>" item with content "<content>"
    Then parents receive sent message
    And other teachers receive sent message
    Examples:
      | message item | content        | num parent | num teacher |
      | text         | Hello world    | 1          | 1           |
      | file         | test-doc.pdf   | 1          | 2           |
      | image        | test-image.jpg | 2          | 2           |

  Scenario Outline: Teacher sends chat items to parents when parents are not present
    Given a chat between "<num parent>" parents and "<num teacher>" teachers
    And "parents" device tokens exist in DB
    And teachers are present
    But parents are not present
    When a teacher sends "<message item>" item with content "<content>"
    Then parents receive notification
    Examples:
      | message item | content        | num parent | num teacher |
      | text         | Hello world    | 1          | 1           |
      | file         | test-doc.pdf   | 1          | 2           |
      | image        | test-image.jpg | 2          | 2           |

    #Done
  Scenario Outline: one parent reading message marks message as read
    Given a chat between "<num parent>" parents and "<num teacher>" teachers
    And parents and teachers are present
    When a teacher sends "<message item>" item with content "<content>"
    And a parent seen the message
    Then teachers see the message has been read
    Examples:
      | message item | content        | num parent | num teacher |
      | text         | Hello world    | 1          | 1           |
      | file         | test-doc.pdf   | 1          | 2           |
      | image        | test-image.jpg | 1          | 2           |

