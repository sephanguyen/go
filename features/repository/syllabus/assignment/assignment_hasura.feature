Feature: Test Assignment Hasura

  Background: 
    Given user create a topic
    # a topic

  Scenario Outline: FindAssignment
    Given a user insert some assignments with that topic id to database
    When user get assignments by call <name>
    Then our system must return <content> correctly

    Examples: 
      | name                   | content                |
      | AssignmentsByTopicIds  | AssignmentsByTopicIds  |
      | AssignmentOne          | AssignmentOne          |
      | AssignmentsMany        | AssignmentsMany        |
      | AssignmentDisplayOrder | AssignmentDisplayOrder |
