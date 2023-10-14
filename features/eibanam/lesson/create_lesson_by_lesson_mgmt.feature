Feature: Create a lesson by lesson management

  Background:
    Given "school admin" logins CMS
    And "teacher" logins Teacher App
    And "student" logins Learner App

  Scenario Outline: School admin can create a lesson with all required fields
    When school admin creates a new lesson with "<start time>", "<end time>", "<teaching medium>", "<teaching method>", teachers, learners, center, media and "<saving option>"
    Then school admin sees the new lesson on Lesson management
#    And school admin sees message "You have created a lesson successfully!" on CMS
#    And teacher sees the new lesson in respective course on Teacher App
#    And student sees the new lesson in lesson list on Learner App

    Examples:
      | start time                | end time                  | teaching medium                | teaching method                   | saving option |
      | 2020-01-02T09:00:00+08:00 | 2020-01-03T09:00:00+08:00 | LESSON_TEACHING_MEDIUM_OFFLINE | LESSON_TEACHING_METHOD_INDIVIDUAL | save one time |