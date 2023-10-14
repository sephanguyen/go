Feature: Export course access paths

    Export subjects masterdata

    Scenario: Export subjects
        Given "school admin" signin system
        And some subjects existed in DB
        When user export subjects
        Then returns subjects in csv with Ok status code
