Feature: Get all conversation of a class
  As a teacher, I want to get all conversation existed of my class

  @wip @deprecated
  Scenario: unauthenticated user get all conversation of a class
    Given a invalid "student" token
    And a ConversationByLessonRequest
    When a user get all conversation of lesson
    Then returns "Unauthenticated" status code

  @wip @deprecated
  Scenario: teacher get all conversation of a class belong to him/her
    Given a EvtLesson with message "CreateLesson"
    When bob send event EvtLesson
    Then tom must create conversation for all lesson

    Given a EvtLesson with message "JoinLesson"
    And a valid "USER_GROUP_STUDENT" id in JoinLesson
    When bob send event EvtLesson
    Then tom must add above user to this lesson conversation

    And hack 1 one message in above conversation

    And a ConversationByLessonRequest
    And a "current lessonID" in ConversationByLessonRequest
    When a user get all conversation of lesson
    Then returns "OK" status code
    And tom must return 1 conversation of lesson

