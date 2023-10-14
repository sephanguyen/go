Feature: Teacher retrieve stream token
  In order for teacher start live stream
  As a teacher
  I need to retrieve a stream token

  Scenario: teacher from same school retrieve lesson token
    Given a teacher from same school with valid lesson
    When teacher join lesson
    Then Bob returns "OK" status code
    And returns valid information for broadcast
    And Tom must record new lesson conversation
    And Tom must record message join lesson of current user
