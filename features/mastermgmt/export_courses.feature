Feature: Export Courses

    Export courses masterdata

    Scenario Outline: Export courses
        Given "school admin" signin system
        And courses existed in DB
        When user export courses
        Then returns courses in csv with Ok status code

