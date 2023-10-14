Feature: Event Class Room

  Background: default manabie resource path
    Given resource path of school "Manabie" is applied

  Scenario: CreateLesson event from Bob

    Given a EvtLesson with message CreateLesson with 2 students
    When bob send event EvtLesson
    Then tom must create conversation for all lesson
    And tom must create conversation member for student in CreateLesson

  Scenario: student cannot join lesson proactively
    Given a lesson conversation with "0" teachers and "0" students
    And a EvtLesson with message "JoinLesson"
    And a valid "USER_GROUP_STUDENT" id in JoinLesson
    When bob send event EvtLesson
    Then tom "must not" add above user to this lesson conversation

  Scenario: teacher can join lesson proactively
    Given a lesson conversation with "1" teachers and "0" students
    And a EvtLesson with message "JoinLesson"
    And a valid "USER_GROUP_TEACHER" id in JoinLesson
    When bob send event EvtLesson
    Then tom "must" add above user to this lesson conversation

  Scenario: student cannot leave conversation
    Given a lesson conversation with "0" teachers and "1" students
    When bob send LeaveLesson for one of previous "student"
    Then tom must not remove student from conversation

  Scenario: teacher can leave lesson proactively
    Given a lesson conversation with "1" teachers and "2" students
    When bob send LeaveLesson for one of previous "teacher"
    Then tom must remove teacher from conversation

  Scenario: teacher can join lesson proactively
    Given a lesson conversation with "0" teachers and "2" students
    When bob send UpdateLesson with "2" new student and without "1" previous students
    Then tom must correctly store only latest students in lesson conversation


  Scenario: teacher end live lesson
    Given a lesson conversation with "0" teachers and "2" students
    And a teacher joins lesson creating new lesson session
    And a second teacher joins lesson without refreshing lesson session
    And students join lesson without refreshing lesson session
    And a EvtLesson with message "EndLesson"
    When bob send event EvtLesson
    Then the "students" in lesson receives "1" message with type "system" with content "CODES_MESSAGE_TYPE_END_LIVE_LESSON"
    And the "second teacher" in lesson receives "1" message with type "system" with content "CODES_MESSAGE_TYPE_END_LIVE_LESSON"


  Scenario: Jprep sync student lesson
    Given a lesson conversation with "0" teachers and "3" students
    When yasuo send EventSyncUserCourse inserting "1" students and deleting "1" prevous student for current lesson
    Then tom must correctly store only latest students in lesson conversation
