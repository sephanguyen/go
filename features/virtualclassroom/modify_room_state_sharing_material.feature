Feature: Modify virtual classroom room sharing material state

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

  Scenario: teacher try to share a video material and after stop sharing
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    When user share a material with type is video in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is video

  Scenario: teacher try to share a PDF material and after stop sharing
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    When user share a material with type is pdf in a virtual classroom session
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is pdf

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user get current material state of a virtual classroom session is pdf
    Then returns "OK" status code

    Given "teacher" signin system
    When user stop sharing material in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is empty
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection
  
  Scenario: teacher try to share an audio material and after stop sharing
    Given "teacher" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    When user share a material with type is audio in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is audio

    Given user signed as student who belong to lesson
    When user join a virtual classroom session
    Then returns "OK" status code
    When user get current material state of a virtual classroom session is audio
    Then returns "OK" status code

    Given "teacher" signin system
    When user stop sharing material in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is empty
    And have a uncompleted log with "2" joined attendees, "3" times getting room state, "2" times updating room state and "0" times reconnection

  Scenario: staff try to share video and then audio material then lastly after stop sharing
    Given "staff granted role school admin" signin system
    And user join a virtual classroom session
    And returns "OK" status code
    When user share a material with type is video in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is video
    
    When user share a material with type is audio in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is audio

    When user stop sharing material in virtual classroom
    Then returns "OK" status code
    And user get current material state of a virtual classroom session is empty