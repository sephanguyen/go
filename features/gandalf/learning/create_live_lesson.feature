Feature: Create live lesson

  Background:
    Given some teacher accounts with school id
    And some student accounts with school id
    And some live courses with school id
    And some medias

  Scenario Outline: User try to create live lesson
    Given user signed as admin
    When user creates live lesson with "<name>", "<start time>", "<end time>", "<brightcove video url>", teachers, courses and learners
    Then returns "OK" status code
    And there is a live lesson with "<name>", "<start time>", "<end time>" and "<brightcove video url>", teachers, courses and learners be created

    Given user signed as teacher
    When user retrieve list lessons by above courses
    Then returns "OK" status code
    And teacher get a live lesson with "<name>", "<start time>", "<end time>" and "<brightcove video url>", teachers, courses and learners be created
    And teacher get a conversation in a room for this lesson with "<name>"

    Examples:
      | name           | start time                | end time                  | brightcove video url                                  |
      | math lesson    | 2020-01-02T09:00:00+08:00 | 2020-01-02T09:00:00+08:00 | https://brightcove.com/account/2/video?videoId=abc123 |
      | physics lesson | 2020-01-02T09:00:00+08:00 | 2020-01-03T09:00:00+08:00 |                                                       |

  Scenario Outline:  User cannot create lesson
    Given user signed as admin
    When user creates lesson with missing "<field>"
    Then user cannot create any lesson
    Examples:
      | field       |
      | lesson name |
      | start time  |
      | end time    |
      | teacher ids |
      | course ids  |
      | student ids |