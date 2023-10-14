Feature: User upsert recurring lesson with some student course duration not alinged with lesson date

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some medias

   Scenario: School admin can create recurring lesson that some student course not aligned with lesson date
    Given user signed in as school admin
    And have some student subscriptions with "<start_at>" and "<end_at>"
    When user creates recurring lesson
    Then returns "OK" status code
    And the recurring lesson was created successfully
    Examples:
      | start_at                   | end_at                       |
      | 2022-07-08T07:00:00+07:00  | 2022-07-17T07:00:00+07:00    | 

    Scenario: School admin can update recurring lesson that some student course not aligned with lesson date
    Given user signed in as school admin
    And have some student subscriptions
    And user have created recurring lesson
    And user changed lesson student info to "<start_at>", "<end_at>"
    When user update selected lesson by saving weekly recurrence
    Then returns "OK" status code
    And the selected lesson & all of the followings were updated
    Examples:
      | start_at                   | end_at                       |  
      | 2022-07-06T07:00:00+07:00  | 2022-07-24T07:00:00+07:00    | 