Feature: User updates (edits) lesson

  Background:
    Given a random number
    And some centers
    And some teacher accounts with school id
    And some student accounts with school id
    And some courses with school id
    And some student subscriptions
    And some medias
    And a lesson

  Scenario Outline: Admin updates lessons
    Given "staff granted role school admin" signin system
    When user updates "<field>" in the existing lesson
    Then returns "OK" status code
    And the lesson is updated
    And Bob must push msg "UpdateLesson" subject "Lesson.Updated" to nats

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