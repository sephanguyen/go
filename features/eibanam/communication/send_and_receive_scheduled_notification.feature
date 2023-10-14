@cms @learner @parent
@communication
@scheduled-notification
@ignore

Feature: Send and receive scheduled notification

  Background:
    Given "school admin" logins CMS
    And school admin has created a student with grade, course and parent info
    And "student" logins Learner App
    And "parent P1" of "student" logins Learner App
    And school admin has created scheduled notification
    And school admin is at "Notification" page on CMS

  Scenario: Send and receive scheduled notification successfully
    When school admin waits for scheduled notification to be sent on time
    Then scheduled notification is sent successfully on CMS
    And "student" receives the scheduled notification in their device
    And "parent P1" receives the scheduled notification in their device

  Scenario: Send and receive scheduled notification successfully after edit sending time
    Given school admin has edited sending time of scheduled notification
    When school admin waits for scheduled notification to be sent on time
    Then scheduled notification is sent successfully on CMS
    And "student" receives the scheduled notification in their device
    And "parent P1" receives the scheduled notification in their device