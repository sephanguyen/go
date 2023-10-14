Feature: Retrieve map lm id and study plan item id by study plan id

    Background: Given a list study plan item
        Given "school admin" logins "CMS"
        And "school admin" has created a content book
        And "school admin" create study plan from the book

    Scenario: User retrieve map lm id and study plan item id
        Given user retrieve map lm id and study plan item id
        Then returns "OK" status code
        And our system must return map lm id and study plan item id correctly
