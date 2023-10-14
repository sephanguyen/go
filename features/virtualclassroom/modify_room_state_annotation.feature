Feature: Modify a virtual classroom session annotation state

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

  Scenario: teacher try to enable and after disable annotation
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user disables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state
    And have a uncompleted log with "2" joined attendees, "2" times getting room state, "3" times updating room state and "0" times reconnection

  Scenario: teacher try disable all annotation after enable
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user disables all annotation learners in a virtual classroom session
    Then returns "OK" status code
    And all annotation state is disable
    And have a uncompleted log with "2" joined attendees, "2" times getting room state, "3" times updating room state and "0" times reconnection

  Scenario: teacher try to stop share pdf after enable annotation
    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is pdf

    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state

    Given "teacher" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user stop sharing material in virtual classroom
    Then returns "OK" status code
    And user get annotation state
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "3" times updating room state and "0" times reconnection

  Scenario: staff tries to enable and after disable annotation
    Given "staff granted role school admin" signin system
    When user join a virtual classroom session
    Then returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    When user enables annotation learners in a virtual classroom session
    Then returns "OK" status code
    And user get annotation state