Feature: Listen to Student Event Logs

    Scenario: learning time event logs
        Given a signed in "student"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time must be "1800s"

    Scenario: learning time event logs by adding with calculated learning time in DB
        Given a signed in "student"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time must be "5400s"
    # 5400s because each of learning_objective logs above has 1800s learning time

    Scenario: learning time event logs without completed events
        Given a signed in "student"
        And student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        When student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time must not be existed

    Scenario: mixing learning_objective event logs with and without completed events
        Given a signed in "student"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        And student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time must be "3600s"
    # 3600s because each of learning_objective logs above has 1800s learning time

    Scenario: ignore learning time if session id is empty
        Given a signed in "student"
        And student inserts a list of learning_objective event logs with session id empty then sleeping "1s"
        And student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time must be "1800s"

    Scenario: "quiz_finished" learning objective completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And "is_finished_quiz" completeness must be "true"
        And total_lo_finished must be "1"

    Scenario: student finishes 3 learning objectives
        Given a signed in "student"

        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a student inserts a list of event logs

        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a student inserts a list of event logs

        And a learning objective is existed in DB
        And a list of quiz_finished event logs

        When a student inserts a list of event logs
        Then waiting for "1s"
        Then returns "OK" status code
        And total_lo_finished must be "3"

    Scenario: student finishes 3 learning objectives and then retry 3rd learning objective
        Given a signed in "student"

        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a student inserts a list of event logs

        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a student inserts a list of event logs

        And a learning objective is existed in DB
        And a list of quiz_finished event logs

        When a student inserts a list of event logs
        Then waiting for "1s"
        Then returns "OK" status code
        And total_lo_finished must be "3"

        When a student retries the last finished learning objective
        Then returns "OK" status code

    Scenario: "quiz_finished" learning objective completeness event logs of student in country other than COUNTRY_VN
        Given a signed in "student"
        And his owned student UUID
        # And the student "country" is "COUNTRY_MASTER"
        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And "is_finished_quiz" completeness must be "true"
        And total_lo_finished must be "1"

    Scenario: "video_finished" learning objective completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of video_finished event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And "is_finished_video" completeness must be "true"
        And total_lo_finished must not be updated

    Scenario: must update first_quiz_correctness if current value is null
        Given a signed in "student"
        And a learning objective is existed in DB

        And a list of quiz_finished event logs with correctness is 5
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs

        And total_lo_finished must be "1"
        And "first_quiz_correctness" completeness must be "5"

    Scenario: must not update first_quiz_correctness if current value is not null
        Given a signed in "student"
        And a learning objective is existed in DB

        And a list of quiz_finished event logs with correctness is 5
        And a student inserts a list of event logs
        And "first_quiz_correctness" completeness must be "5"

        And a list of video_finished event logs
        And a student inserts a list of event logs

        And a list of quiz_finished event logs with correctness is 10
        When a student inserts a list of event logs
        Then returns "OK" status code
        And "first_quiz_correctness" completeness must be "5"

    Scenario: must update highest_quiz_score
        Given a signed in "student"
        And a learning objective is existed in DB

        And a list of quiz_finished event logs with correctness is 5
        And a student inserts a list of event logs
        And "first_quiz_correctness" completeness must be "5"
        And "highest_quiz_score" completeness must be "5"

        And a list of quiz_finished event logs with correctness is 10
        When a student inserts a list of event logs
        And "first_quiz_correctness" completeness must be "5"
        Then "highest_quiz_score" completeness must be "10"

    Scenario: "study_guide_finished" learning objective completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of study_guide_finished event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And "is_finished_study_guide" completeness must be "true"
        And total_lo_finished must not be updated

    Scenario: mixing learning objective completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a list of video_finished event logs
        And a list of study_guide_finished event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And total_lo_finished must be "1"
        And "is_finished_quiz" completeness must be "true"
        And "is_finished_video" completeness must be "true"
        And "is_finished_study_guide" completeness must be "true"

    Scenario: mixing learning_objective and learning completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a list of video_finished event logs
        And a list of study_guide_finished event logs
        And a student inserts a list of event logs
        And a list of learning_objective event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And total_lo_finished must be "1"
        And total_learning_time must be "1800s"
    # 1800s because the learning_objective logs above has 1800s learning time

    Scenario: student finishes tutorial lo
        Given a signed in "student"
        And his owned student UUID
        And student finishes tutorial lo
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs

    Scenario: learning time event logs v2
        Given a signed in "student"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time v2 must be "1800s"

    Scenario: learning time event logs by adding with calculated learning time in DB v2
        Given a signed in "student"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time v2 must be "5400s"
    # 5400s because each of learning_objective logs above has 1800s learning time

    Scenario: mixing learning_objective event logs with and without completed events v2
        Given a signed in "student"
        And student inserts a list of learning_objective event logs then sleeping "1s"
        And student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time v2 must be "3600s"
    # 3600s because each of learning_objective logs above has 1800s learning time

    Scenario: ignore learning time if session id is empty
        Given a signed in "student"
        And student inserts a list of learning_objective event logs with session id empty then sleeping "1s"
        And student inserts a list of learning_objective event logs without completed event then sleeping "1s"
        When student inserts a list of learning_objective event logs then sleeping "1s"
        Then returns "OK" status code
        And total_learning_time v2 must be "1800s"

    Scenario: mixing learning_objective and learning completeness event logs
        Given a signed in "student"
        And a learning objective is existed in DB
        And a list of quiz_finished event logs
        And a list of video_finished event logs
        And a list of study_guide_finished event logs
        And a student inserts a list of event logs
        And a list of learning_objective event logs
        When a student inserts a list of event logs
        Then returns "OK" status code
        And Eureka must record all student's event logs
        And total_lo_finished must be "1"
        And total_learning_time v2 must be "1800s"
    # 1800s because the learning_objective logs above has 1800s learning time
