Feature: [Not function/integration test] [DB test only] Database admin add record to student study plan
    The student study plan of each student must link only one master study plan

Background: 
    Given a course study plan created
Scenario Outline: the database admin create some student study plans and some duplicate
    When the database admin create some study plans for a which have "<status>" master study plan
    Then our database have to handle <"case"> correctly
    Examples:
        |   status       |      case        |
        |  no duplicate  |  no duplicate    |
        |  duplicate     |  duplicate       |