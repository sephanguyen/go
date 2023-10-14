@runsequence
Feature: User update recurring lesson
  
  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And have some classrooms
    And have zoom account owner

  # general info included (teaching medium, teaching method, teacher, course, class, student, material)

  Scenario: User update lesson general info of recurring lesson
    Given user signed in as school admin
    And user have created recurring lesson
    And user changed lesson general info to "<end_date>"
    When user update selected lesson by saving weekly recurrence
    Then returns "OK" status code
    And the selected lesson & all of the followings were updated
    And the locked lessons were not updated "<start_time>", "<end_time>", "none_location", "general_infor" or deleted
    And student and teacher name must be updated correctly
    Examples:
      | end_date                  |
      | 2022-07-31T08:30:00+07:00 |
      | 2022-08-14T08:30:00+07:00 |
  # case 1 : no change end_date
  # case 2 : change end_date to add two lesson
  # case 2 : change end_date to delete one lesson

  Scenario Outline: User update lesson time of recurring lesson
    Given user signed in as school admin
    And user have created recurring lesson
    And the lesson has start_time is "<locked_at>" will locked
    And user changed lesson time to "<start_time>" , "<end_time>" and "<end_date>"
    When user update selected lesson by saving weekly recurrence
    Then returns "OK" status code
    And the selected lesson & all of the followings were updated and link with new recurring chain
    And end date of old chain was updated
    And the locked lessons were not updated "<start_time>", "<end_time>", "none_location", "none_general_infor" or deleted
    Examples:
      | start_time                | end_time                  | end_date                  | locked_at                 |
      | 2022-07-23T07:00:00+07:00 | 2022-07-23T08:30:00+07:00 | 2022-07-31T08:30:00+07:00 |                           |
      | 2022-07-23T07:00:00+07:00 | 2022-07-23T08:30:00+07:00 | 2022-08-14T08:30:00+07:00 |                           |
      | 2022-07-23T07:00:00+07:00 | 2022-07-23T08:30:00+07:00 | 2022-07-24T08:30:00+07:00 |                           |
      | 2022-07-16T07:00:00+07:00 | 2022-07-16T08:30:00+07:00 | 2022-07-24T08:30:00+07:00 | 2022-07-30T02:00:00+00:00 |
      | 2022-07-16T07:00:00+07:00 | 2022-07-16T08:30:00+07:00 | 2022-07-24T08:30:00+07:00 | 2022-07-24T08:30:00+07:00 |

  # case 1 : no change end_date
  # case 2 : change end_date to add two lesson
  # case 3 : change end_date to delete one lesson

  Scenario Outline: User update lesson location of recurring lesson
    Given user signed in as school admin
    And user have created recurring lesson
    And user changed lesson location to "<location>" and "<end_date>"
    When user update selected lesson by saving weekly recurrence
    Then returns "OK" status code
    And the selected lesson & all of the followings were updated and link with new recurring chain
    And end date of old chain was updated
    And the locked lessons were not updated "none_start_time", "none_end_time", "<location>", "none_general_infor" or deleted
    Examples:
      | location | end_date                  |
      | 1        | 2022-07-31T08:30:00+07:00 |
      | 1        | 2022-08-14T08:30:00+07:00 |
      | 1        | 2022-07-24T08:30:00+07:00 |

Scenario Outline: User update lesson time of recurring lesson with only this lesson
  Given user signed in as school admin
  And user have created recurring lesson
  And user changed lesson time to "<start_time>" , "<end_time>" and "<end_date>"
  When user update selected lesson by saving only this
  Then returns "OK" status code
  And the selected lesson was updated
  And the selected lesson still keep in old chain
  Examples:
    | start_time                | end_time                  | end_date                  |
    | 2022-07-23T07:00:00+07:00 | 2022-07-23T08:30:00+07:00 | 2022-07-31T08:30:00+07:00 |

Scenario Outline: User update lesson location of recurring lesson with only this lesson
  Given user signed in as school admin
  And user have created recurring lesson
  And user changed lesson location to "<location>" and "<end_date>"
  When user update selected lesson by saving only this
  Then returns "OK" status code
  And the selected lesson was updated
  And the selected lesson was updated in new scheduler
  Examples:
    | location | end_date                  |
    | 1        | 2022-07-31T08:30:00+07:00 |

Scenario Outline: School admin can save draft/publish lesson in recurring chain
  Given user signed in as school admin
  When user creates "<lesson_status>" recurring lesson with "<missing_fields>"
  And user changed lesson general info
  Then user save "<saving_type>" selected lesson by saving weekly recurrence
  And the selected lesson & all of the followings were updated
  Examples:
    | lesson_status | missing_fields        | saving_type |
    | draft         | none                  | published   |
    | draft         | students and teachers | published   |
    | published     | none                  | draft       |  

Scenario Outline: User update recurring lesson with only this lesson and available date type
  Given user signed in as school admin
  And user have created recurring lesson
  And a date "2022-08-01T00:00:00+07:00", location "1", date type "regular", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-08T00:00:00+07:00", location "1", date type "spare", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-15T00:00:00+07:00", location "1", date type "seasonal", open time "8:30", status "published", resource path "-2147483648" are existed in DB
  And a date "2022-08-22T00:00:00+07:00", location "1", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-29T00:00:00+07:00", location "1", date type "regular", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And user changed lesson time to "<start_time>", "<end_time>", "<end_date>" and location "<location>"
  When user update selected lesson by saving only this
  Then returns "OK" status code
  And returns some lesson dates
  And recurring lessons will include "<expected dates>" and skip "<exclude date>"
  Examples:
    | start_time                | end_time                  | end_date                  | location  | expected dates          | exclude date                      |
    | 2022-08-01T07:00:00+07:00 | 2022-08-01T08:30:00+07:00 | 2022-08-31T08:30:00+07:00 | 1         | 2022-08-01              | 2022-08-08,2022-08-15,2022-08-22  |

Scenario Outline: User update recurring lesson with recurring chain lesson and available date type
  Given user signed in as school admin
  And user have created recurring lesson
  And a date "2022-08-01T00:00:00+07:00", location "2", date type "regular", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-08T00:00:00+07:00", location "2", date type "spare", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-15T00:00:00+07:00", location "2", date type "seasonal", open time "8:30", status "published", resource path "-2147483648" are existed in DB
  And a date "2022-08-22T00:00:00+07:00", location "2", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-31T00:00:00+07:00", location "2", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-17T00:00:00+07:00", location "2", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
  And a date "2022-08-29T00:00:00+07:00", location "2", date type "regular", open time "8:30", status "draft", resource path "-2147483648" are existed in DB
  And user changed lesson time to "<start_time>", "<end_time>", "<end_date>" and location "<location>"
  When user update selected lesson by saving weekly recurrence
  Then returns "OK" status code
  And returns some lesson dates
  And recurring lessons will include "<expected dates>" and skip "<exclude date>"
  # case 1: change start_date with date type = 'regular'
  # case 2: change start_date to new start date has no date info
  # case 3: change start_date with date type = 'spare'
  Examples:
    | start_time                | end_time                  | end_date                  | location  | expected dates                    | exclude date                      |
    | 2022-08-01T07:00:00+07:00 | 2022-08-01T08:30:00+07:00 | 2022-08-31T08:30:00+07:00 | 2         | 2022-08-01,2022-08-29             | 2022-08-08,2022-08-15,2022-08-22  |
    | 2022-08-01T04:00:00+07:00 | 2022-08-01T12:30:00+07:00 | 2022-08-31T12:30:00+07:00 | 2         | 2022-08-01,2022-08-29             | 2022-08-08,2022-08-15,2022-08-22  |
    | 2022-08-10T07:00:00+07:00 | 2022-08-10T08:30:00+07:00 | 2022-08-31T08:30:00+07:00 | 2         | 2022-08-10,2022-08-24             | 2022-08-31                        |
    | 2022-08-08T07:00:00+07:00 | 2022-08-08T08:30:00+07:00 | 2022-09-09T08:30:00+07:00 | 2         | 2022-08-08,2022-08-29,2022-09-05  | 2022-08-15,2022-08-22             |
    | 2022-08-08T04:00:00+07:00 | 2022-08-08T04:30:00+07:00 | 2022-09-09T04:30:00+07:00 | 2         | 2022-08-08,2022-08-29,2022-09-05  | 2022-08-15,2022-08-22             |

Scenario Outline: User update recurring lesson classroom with only this lesson
  Given user signed in as school admin
  And an existing recurring lesson with classroom
  And user changed lesson with "existing" classroom
  When user update selected lesson by saving only this
  Then returns "OK" status code
  And the selected lesson classroom is "updated"
  And the other lessons classroom are "not updated"

Scenario Outline: User update recurring lesson classroom with recurring chain lesson
  Given user signed in as school admin
  And an existing recurring lesson with classroom
  And user changed lesson with "existing" classroom
  When user update selected lesson by saving weekly recurrence
  Then returns "OK" status code
  And the selected lesson classroom is "updated"
  And the other lessons classroom are "updated"
 @quarantined
  Scenario: User update lesson zoom info of recurring lesson
    Given user signed in as school admin
    And user have created recurring lesson with zoom info
    And user changed lesson general info to "<end_date>"
    When user update selected lesson by saving weekly recurrence
    Then returns "OK" status code
    Examples:
      | end_date                  |
      | 2023-02-20T08:30:00+07:00 |
