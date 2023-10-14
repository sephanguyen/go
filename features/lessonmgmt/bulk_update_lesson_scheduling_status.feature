Feature: User bulk updates scheduling status

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: School admin bulk update lessons scheduling status
    Given user signed in as school admin
    And a list of existing lessons with scheduling status "<status>" and teaching method "<teaching-method>"
    When user bulk updates scheduling status with action "<action>"
    Then returns "OK" status code
    And the lessons scheduling status are updated correctly to "<expected-status>"
   
    Examples:
      | status    | teaching-method | action  | expected-status |
      # correct cases: the status will be updated accordingly to the action
      | published | individual      | cancel  | canceled        |
      | completed | individual      | cancel  | canceled        |
      | draft     | individual      | publish | published       |
      | draft     | group           | publish | published       |
      # incorrect cases: the status will be kept the same as "<status>"
      | draft     | individual      | cancel  | draft           |
      | canceled  | individual      | cancel  | canceled        |
      | published | individual      | publish | published       |
      | completed | individual      | publish | completed       |
      | canceled  | individual      | publish | canceled        |

