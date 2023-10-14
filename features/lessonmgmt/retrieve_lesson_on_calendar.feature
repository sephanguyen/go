Feature: Retrieve lessons on calendar view

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some classrooms
    And "enable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_SwitchNewDBConnection"

  Scenario Outline: User can retrieve lessons on calendar view
    Given user signed in as school admin
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
    And user creates a set of lessons for "<lesson_dates>"
    When user retrieves lessons on "<calendar_view>" from "<from_date>" to "<to_date>"
    Then returns "OK" status code
    And lessons retrieved are within the date range "<from_date>" to "<to_date>"
    And the lessons first and last date matches with "<from_date>" and "<to_date>"

    Examples:
      | lesson_dates          | calendar_view | from_date  | to_date    |
      | 2022-09-06            | DAILY         | 2022-09-06 | 2022-09-06 |
      | 2022-09-12,2022-09-18 | WEEKLY        | 2022-09-12 | 2022-09-18 |
      | 2022-08-29,2022-10-02 | MONTHLY       | 2022-08-29 | 2022-10-02 |

    Scenario Outline: User can retrieve lessons on calendar view with filter
    Given user signed in as school admin
    And a list of lessons are existing
    When user retrieves lessons on "<calendar_view>" with filter "<students>", "<teachers>", "<courses>", "<classes>", "<none_assigned_teacher>"
    Then returns "OK" status code
    And lessons retrieved on calendar with filter are correct

    Examples:
        | calendar_view | students            | teachers            | courses  | classes         | none_assigned_teacher |
        | DAILY         | student-1,student-2 | teacher-1,teacher-2 | course-1 |                 | true                  |
        | DAILY         | student-1           |                     |          |                 | false                 |
        | WEEKLY        |                     | teacher-1           | course-1 | class-1,class-2 | false                 |
        | MONTHLY       | student-1           | teacher-1           | course-1 | class-1         | false                 |

    Scenario Outline: User can retrieve lessons on calendar view with filter and Unleash turned on
    Given user signed in as school admin
    And a list of lessons are existing
    When user retrieves lessons on "<calendar_view>" with filter "<students>", "<teachers>", "<courses>", "<classes>", "<none_assigned_teacher>"
    Then returns "OK" status code
    And lessons retrieved on calendar with filter are correct
    Examples:
        | calendar_view | students            | teachers            | courses  | classes         | none_assigned_teacher |
        | DAILY         | student-1,student-2 | teacher-1,teacher-2 | course-1 |                 | true                  |
        | DAILY         | student-1           |                     |          |                 | false                 |
        | DAILY         | student-1           |                     |          |                 | true                  |
        | WEEKLY        |                     | teacher-1           | course-1 | class-1,class-2 | true                  |
        | MONTHLY       | student-1           | teacher-1           | course-1 | class-1         | true                  |
