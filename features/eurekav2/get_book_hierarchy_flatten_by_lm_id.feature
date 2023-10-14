Feature: Get a BookHierarchyFlatten by learningMaterialID

  Background:
    Given a signed in "school admin"
    And user adds a simple book content
    And user adds some learning materials to topic of the book

  Scenario: Get a valid BookHierarchyFlatten
    When user gets a "existing" book hierarchy flatten
    Then returns "OK" status code
    And returns correct book hierarchy flatten of that learning material

  Scenario: Get an invalid BookHierarchyFlatten
    When user gets a "not-existing" book hierarchy flatten
    Then returns "NotFound" status code
