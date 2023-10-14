Feature: Authentication when user upsert single quiz
    Background:
        Given a signed in "school admin"
        And learning objective belonged to a topic
    
    Scenario Outline: User upsert single quiz
        Given a signed in "<role>"
        When user upsert a valid "<quiz>" single quiz with role
        Then returns "<status>" status code

        Examples:
            | role           | status           |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | center lead    | PermissionDenied |
            | parent         | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |
