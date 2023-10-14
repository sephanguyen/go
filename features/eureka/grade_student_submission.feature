@quarantined
Feature: Grade student submission

    Background: Grade student submission
        Given "school admin" logins "CMS"
        And "student" logins "Learner App"
        And "teacher" logins "Teacher App"
        And "school admin" create a study plan with book have an assignment
        And "school admin" add student to course

    Scenario: Grade student submission
        Given "student" do assignment
        When "teacher" grade submission with status returned
        Then our system returns "OK" status code
        And notification has been stored correctly

    Scenario: Grade student submission when student null
        Given "student" do assignment
        And update student school to null
        When "teacher" grade submission with status returned
        Then our system returns "OK" status code
        And notification has not been stored