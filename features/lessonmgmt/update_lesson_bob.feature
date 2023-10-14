Feature: User updates (edits) lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And a form's config for "individual lesson report" feature with school id

  Scenario Outline: School admin updates a lesson
    Given user signed in as school admin
    And an existing lesson
    When user updates "<field>" in the lesson
    Then returns "OK" status code
    And the lesson was updated
    And Bob must push msg "UpdateLesson" subject "Lesson.Updated" to nats
    And student and teacher name must be updated correctly
    
    Examples:
      | field             |
      | start time        |
      | end time          |
      | center id         |
      | teacher ids       |
      | student info list |
      | teaching medium   |
      | teaching method   |
      | material info     |
      | all fields        |

  Scenario Outline: Teacher updates a lesson
    Given user signed in as teacher
    And an existing lesson
    When user updates "<field>" in the lesson
    Then returns "OK" status code
    And the lesson was updated
    And Bob must push msg "UpdateLesson" subject "Lesson.Updated" to nats
    And student and teacher name must be updated correctly

    Examples:
      | field             |
      | start time        |
      | end time          |
      | center id         |
      | teacher ids       |
      | student info list |
      | teaching medium   |
      | teaching method   |
      | material info     |
      | all fields        |

  Scenario: School admin cannot update lessons with incorrect time
    Given user signed in as school admin
    And an existing lesson
    When user updates the lesson with start time later than end time
    Then returns "Internal" status code
#   And the lesson is not updated

  Scenario Outline: School admin cannot update lessons with empty required fields
    Given user signed in as school admin
    And an existing lesson
    When user updates the lesson with missing "<field>"
    Then returns "Internal" status code
#   And the lesson is not updated

    Examples:
      | field             |
      | center id         |
      | start time        |
      | end time          |
      | teacher ids       |

  Scenario: School admin can update a lesson with "group teaching method" with all required fields
    Given user signed in as school admin
    And a class with id prefix "<prefix-class-id>" and a course with id prefix "<prefix-course-id>"
    And user creates a new lesson with "group" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    When user updates "<field>" in the lesson
    Then returns "OK" status code
    And the lesson was updated
    Examples:
      | field             | prefix-class-id    | prefix-course-id    |
      | start time        | bdd-test-class-id- | bdd-test-course-id- |
      | end time          | bdd-test-class-id- | bdd-test-course-id- |
      | center id         | bdd-test-class-id- | bdd-test-course-id- |
      | teacher ids       | bdd-test-class-id- | bdd-test-course-id- |
      | student info list | bdd-test-class-id- | bdd-test-course-id- |
      | teaching medium   | bdd-test-class-id- | bdd-test-course-id- |
      | teaching method   | bdd-test-class-id- | bdd-test-course-id- |
      | material info     | bdd-test-class-id- | bdd-test-course-id- |
      | all fields        | bdd-test-class-id- | bdd-test-course-id- |

  Scenario: School admin can edit lesson by saving draft from published lesson
    Given user signed in as school admin
    And user creates a "<lesson_status>" lesson with "<missing_fields>" in "<service>"
    And admin create a lesson report
    When user edit lesson by saving "<saving_type>" in "<service>"
    Then returns "OK" status code
    And the lesson was updated
    And lesson report state is "<state>"
   Examples:
      | lesson_status  | missing_fields    |  saving_type  | state       | service |
      | published      | none              |  draft        | deleted     | bob     |
      | draft          | none              |  published    | not deleted | bob     |
      | draft          | student info list |  draft        | deleted     | bob     |