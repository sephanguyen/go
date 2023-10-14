Feature: Export Grades

    Export grades masterdata

    Scenario Outline: Export grades
        Given "school admin" signin system
        And some grades existed in DB
        When user export grades
        Then returns grades in csv with Ok status code
