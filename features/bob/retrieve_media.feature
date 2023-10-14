Feature: Retrieve media
    To retrieve media

    Scenario: student retrieve media by ids
        Given student has multiple media
        When student retrieve media by ids 
        Then Bob must return all media