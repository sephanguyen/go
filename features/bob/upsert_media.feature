Feature: Upsert media
    To upload media

    Scenario: student upsert media
        Given a signed in student
        When user upsert valid media list
        Then Bob must record all media list
