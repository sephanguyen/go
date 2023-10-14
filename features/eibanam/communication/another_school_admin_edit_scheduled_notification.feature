@cms
@communication
@ignore

Feature: Another school admin edits scheduled notification

  Background:
    Given "school admin 1" logins CMS
    And "school admin 1" has created student S1 with parent P1 info
    And "school admin 1" has created a scheduled notification
    And "school admin 2" logins CMS

  Scenario Outline: Another school admin can update <field> of scheduled notification successfully with <button> button
    Given "school admin 2" has opened editor full-screen dialog of scheduled notification
    When "school admin 2" edits "<field>" of scheduled notification
    And "school admin 2" clicks "<button>" button
    Then "school admin 2" sees updated scheduled notification on CMS
    And "school admin 2" sees name of composer updated to "school admin 2"
    And "school admin 1" sees name of composer updated to "school admin 2"
    Examples:
      | field           | button                       |
      | Title           | 1 of [Save schedule, Close ] |
      | Content         | 1 of [Save schedule, Close ] |
      | Date            | 1 of [Save schedule, Close ] |
      | Time            | 1 of [Save schedule, Close ] |
      | Grade           | 1 of [Save schedule, Close ] |
      | Course          | 1 of [Save schedule, Close ] |
      | Recipient email | 1 of [Save schedule, Close ] |
      | Type filter     | 1 of [Save schedule, Close ] |
      | All fields      | 1 of [Save schedule, Close ] |