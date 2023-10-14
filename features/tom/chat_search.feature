Feature: List conversation

  Background: Students chat and parents chat are created
    Given a new school is created
    And a valid "school admin" token
    And a teacher account in db

#  @blocker
  Scenario Outline: Pagination
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    And chat pagination 10 item
    And all chat is returned
    When pagination by offset of "<pagination item>" limit 100
    Then pagination return "<final pagination items>" number of items
    Examples:
      | pagination item | final pagination items |
      | next page       | 0                      |
      | first item      | num conversations - 1  |


  @throttle
  Scenario Outline: Unreplied/replied filter
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    And some student's chat has new message from student
    And some parents' chat has new message from parent
    And teacher replies to some of those chat
    And filter chats with type message type "<message type>" only
    Then only chats with status "<message type>" are returned
    Examples:
      | message type |
      | Replied      |
      | Unreplied    |
      | All          |

  Scenario: Teacher replying makes chat disappear in unreplied filter result
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    And a result of unreplied chats filter
    Then teacher replies to an unreplied chat
    And filter chats with type message type "Unreplied" only
    Then only chats with status "Unreplied" are returned including previous replied chat

  Scenario Outline: Teacher filter student/parent chats only
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    And filter chats with type contact type "<contact type>" only
    Then only chats with type "<contact type>" are returned
    And returned chats have correct student ids
    Examples:
      | contact type |
      | All          |
      | Student      |
      | Parent       |

  Scenario: Teacher filter 1 course
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When filter by a course id
    Then chats that belong to that course are returned

  Scenario: Teacher filter multiple courses
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When filter by multiple courses
    Then chats that belong to those courses are returned

  @throttle
  Scenario Outline: Teacher filter multiple courses and combine with message type, contact type
    Given "english" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When filter by multiple courses
    And filter chats with type message type "<message type>" only
    And filter chats with type contact type "<contact type>" only
    Then chats with type "<message type>", "<contact type>" that belong to those courses are returned
    Examples:
      | message type | contact type |
      | All          | All          |
      | All          | Student      |
      | All          | Parent       |
      | Unreplied    | All          |
      | Unreplied    | Student      |
      | Unreplied    | Parent       |

    # Done
    Scenario Outline: Search by partial name
      Given "<language form>" language is used
      And students and each has from 1 to 1 parents randomly
      And each student joined some courses at random
      When use partial names of a student to search
      Then chats with name including the partial names are returned
      Examples:
        | language form |
        | english       |

    # Done
  Scenario Outline: student name updated to student/parent chat elasticsearch
    Given "<language form>" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When a student name is updated with event "EvtUserInfo"
    And use updated name of student to search
    Then student and parent chats with updated names are returned
    Examples:
      | language form |
      | english       |


  Scenario Outline: Search by full name
    Given "<language form>" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When use full names of a student to search
    Then chats with name including the full names are returned
    Examples:
      | language form |
      | english       |


  Scenario Outline: Search by full name with course that student does not belong to
    Given "<language form>" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    When use full names of a student to search
    And filter by a course that the student does not belong to
    Then nothing is returned
    Examples:
      | language form |
      | english       |

  @throttle
  Scenario Outline: Search student/parent chats by full name and filter by multiple courses and replied status
    Given "<language form>" language is used
    And students and each has from 1 to 1 parents randomly
    And each student joined some courses at random
    And some student's chat has new message from student
    And some parents' chat has new message from parent
    And teacher replies to some of those chat
    When filter by multiple courses
    And filter chats with type message type "<message type>" only
    And filter chats with type contact type "<contact type>" only
    When use full names of a student to search
    Then chats with status "<message type>" with type "<contact type>" belonging to those courses and having name including the name of the student are returned
    Examples:
      | contact type | message type | language form |
      | Parent       | Replied      | english       |
      | Student      | Unreplied    | hiragana      |
      | Student      | Unreplied    | katakana      |
      | Student      | Unreplied    | kanji         |
