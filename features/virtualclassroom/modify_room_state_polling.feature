Feature: Modify a virtual classroom session polling state

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing a virtual classroom session
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: teacher try start polling and after stop, end it
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start polling with "<number_options>" options and "<number_correct>" correct answers in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user submit the answer "A,B" for polling
    Then returns "OK" status code
    And user get polling answer state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user stop polling in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session stopped

    When user "start" share polling in a virtual classroom session
    Then returns "OK" status code
    Then user get current polling state of a virtual classroom session "start" share polling

    When user "stop" share polling in a virtual classroom session
    Then returns "OK" status code
    Then user get current polling state of a virtual classroom session "stop" share polling

    When user end polling in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session is empty

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start polling in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user submit the answer "B,C" for polling
    Then returns "OK" status code
    And user get polling answer state
    And have a uncompleted log with "4" joined attendees, "8" times getting room state, "8" times updating room state and "0" times reconnection

    Examples:
      | number_options | number_correct |
      | 2              | 0              |
      | 5              | 0              |
      | 10             | 0              |
      | 2              | 2              |
      | 10             | 10             |

  Scenario: teacher try start polling => end lesson => re-start lesson => polling is reset
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start polling with "<number_options>" options and "<number_correct>" correct answers in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started
    Then user end the live lesson
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session is empty

  Scenario: staff tries to start and stop polling
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user start polling with "3" options and "1" correct answers in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session started
    When user stop polling in a virtual classroom session
    Then returns "OK" status code
    And user get current polling state of a virtual classroom session stopped