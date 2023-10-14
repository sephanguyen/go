Feature: The student's learning material completion time

  Scenario Outline: The student's learning material completion time
    Given a valid "<event type>" student event log in DB at "<insert time>"
    And a valid "<update event>" student event log previous at "<update time>"
    When <student_event_log>checking completed time of learning material
    Then <student_event_log>a valid "<completed time>" of learning material
    Examples:
      | event type | insert time          | update event | update time          | completed time       |
      | completed  | 2022-10-17T07:30:00Z | exited       | 2022-10-17T08:00:00Z | 2022-10-17T08:00:00Z |
      | started    | 2022-10-16T06:30:00Z | exited       | 2022-10-16T07:00:00Z | 2022-10-16T07:00:00Z |
      | started    | 2022-10-15T06:30:00Z | paused       | 2022-10-15T07:00:00Z |                      |

  Scenario Outline: The student submit exam lo
    Given <student_event_log>a valid exam lo in DB
    And <student_event_log>a student "<submit>" exam lo
    When <student_event_log>checking completed time of learning material
    Then <student_event_log>a valid "<completed time>" of learning material
    Examples:
      | submit    | completed time       |
      | doesn't   |                      |
      | submitted | 2023-01-03T08:00:00Z |
