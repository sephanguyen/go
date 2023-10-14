Feature: List submissions v2
    #retrieve submissions of student mapped with those locations
    Background: students submit assignments
        Given a valid "teacher" token
        And some students added to course in some valid locations
        And students are assigned assignments in study plan
        And students submit their assignments

    Scenario: teacher listing submissions with locations
        When "teacher" list submissions using v2 with valid locations
        Then returns "OK" status code
        And our system must returns list submissions correctly

    Scenario: teacher listing submissions with other locations
        When "teacher" list submissions using v2 with invalid locations
        Then returns "OK" status code
        And our system must returns list submissions is empty

    Scenario: teacher listing submissions with other locations
        When "teacher" list submissions using v2 with some valid locations and some invalid locations
        Then returns "OK" status code
        And our system must returns list submissions correctly

    Scenario: teacher listing submissions with locations and course is null
        When "teacher" list submissions using v2 with valid locations and course is null
        Then returns "OK" status code
        And our system must returns list submissions correctly

    Scenario: teacher listing submissions with locations and course is null
        Given a list submissions of students with random locations
        When "teacher" list submissions using v2 with valid locations and course is null
        Then returns "OK" status code
        And our system must returns list submissions correctly

    Scenario: teacher listing submissions with locations
        Given a student expired in course
        When "teacher" list submissions using v2 with valid locations
        Then returns "OK" status code
        And our system must returns list submissions correctly
