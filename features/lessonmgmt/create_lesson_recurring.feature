Feature: User creates recurring lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And have some classrooms

  Scenario: School admin can create recurring lesson
    Given user signed in as school admin
    And existing lesson event subscriber 
    When user creates recurring lesson
    Then returns "OK" status code
    And the recurring lesson was created successfully
    And Lessonmgmt must push msg "CreateLessons" subject "Lesson.Created" to nats

   Scenario: School admin can create recurring lesson with missing some info
    Given user signed in as school admin
    When user creates "<lesson_status>" recurring lesson with "<missing_fields>"
    Then returns "OK" status code
    And the recurring lesson was created successfully
   Examples:
      | lesson_status  | missing_fields         | 
      | draft          | none                   | 
      | draft          | students               | 
      | draft          | teachers               | 
      | draft          | students and teachers  |
      | published      | students               |

   Scenario: School admin can't create recurring lesson missing some info
    Given user signed in as school admin
    When user creates "<lesson_status>" recurring lesson with "<missing_fields>"
    Then returns "Internal" status code
    Examples:
      | lesson_status   | missing_fields         |
      | published       | teachers               | 
      | published       | students and teachers  | 

    Scenario: School admin can create a lesson and with day types
      Given user signed in as school admin
      And a date "2022-08-01T00:00:00+07:00", location "1", date type "regular", open time "07:30", status "draft", resource path "-2147483648" are existed in DB
      And a date "2022-08-08T00:00:00+07:00", location "1", date type "spare", open time "07:30", status "draft", resource path "-2147483648" are existed in DB
      And a date "2022-08-15T00:00:00+07:00", location "1", date type "seasonal", open time "07:30", status "draft", resource path "-2147483648" are existed in DB
      And a date "2022-08-17T00:00:00+07:00", location "1", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
      And a date "2022-08-22T00:00:00+07:00", location "1", date type "closed", open time "", status "draft", resource path "-2147483648" are existed in DB
      And a date "2022-08-29T00:00:00+07:00", location "1", date type "regular", open time "07:30", status "draft", resource path "-2147483648" are existed in DB
      When user creates recurring lesson with "<lesson_date>" and "<location>" until "<end_date>"
      Then returns "OK" status code
      And returns some lesson dates
      And recurring lessons will include "<expected_dates>" and skip "<skipped_dates>"
      Examples:
        | lesson_date  | location  | end_date    | expected_dates        | skipped_dates                    |
        | 2022-08-01   | 1         | 2022-08-31  | 2022-08-01,2022-08-29 | 2022-08-08,2022-08-15,2022-08-22 |
        | 2022-08-02   | 1         | 2022-08-08  | 2022-08-02            |                                  |
        | 2022-08-03   | 1         | 2022-08-18  | 2022-08-03,2022-08-10 | 2022-08-17                       |

  Scenario: School admin can create recurring lesson with classrooms
    Given user signed in as school admin
    When user creates recurring lesson with "<record_state>" classrooms
    Then returns "<status>" status code
    And the recurring lesson is "<create_status>"
    And the classrooms are "<record_state>" in the recurring lesson
    Examples:
      | record_state  | status    | create_status |
      | existing      | OK        | created       |
      | not existing  | Internal  | not created   |