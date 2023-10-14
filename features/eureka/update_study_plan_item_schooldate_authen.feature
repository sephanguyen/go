Feature: authen update school date study plan item

    Background: prepare content book and studyplan belongs to 1 student
        Given "school admin" logins "CMS"
        And "student" logins "Learner App"
        And "school admin" has created a content book
        And "school admin" has created a studyplan exact match with the book content for student

    Scenario Outline: Update school date with valid request
        Given a signed in "<role>"
        When user update school date with valid request
        Then returns "<status code>" status code
        Examples:
            | role         | status code      |
            | parent       | PermissionDenied |
            | student      | PermissionDenied |
            | teacher      | OK               |
            | hq staff     | OK               |
            | school admin | OK               |
# | center lead     | OK               |
# | center staff    | OK               |
# | center manager  | OK               |