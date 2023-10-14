Feature: Modify live lesson room annotation state

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing live lesson

  Scenario: teacher try to enable and after disable annotation
    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user share a material with type is pdf in live lesson room
    Then returns "OK" status code
    When user enables annotation learners in the live lesson room
    Then returns "OK" status code
    And user get annotation state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user disables annotation learners in the live lesson room
    Then returns "OK" status code
    And user get annotation state
    And have a uncompleted log with "2" joined attendees, "2" times getting room state, "3" times updating room state and "0" times reconnection

  Scenario: teacher try to stop share pdf after enable annotation
    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user share a material with type is pdf in live lesson room
    Then returns "OK" status code
    And user get current material state of live lesson room is pdf

    When user enables annotation learners in the live lesson room
    Then returns "OK" status code
    And user get annotation state

    Given user signed in as teacher
    When user join live lesson
    Then returns "OK" status code
    When user stop sharing material in live lesson room
    Then returns "OK" status code
    And user get annotation state
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "3" times updating room state and "0" times reconnection
