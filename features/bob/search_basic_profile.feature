Feature: Seach basic profile

    With input are a list user id and search_text(name)
    I want to retrieve all basic profile satisfy the input.

    Scenario: retrieve with the search_text is nil
        Given a list user valid in db
        When search basic profile "ids" filter
        Then returns a list basic profile correctly
    Scenario: retrieve basic profile
        Given a list user valid in db
            And update a student name
        When search basic profile "full" filter
        Then returns a list basic profile according search_text


