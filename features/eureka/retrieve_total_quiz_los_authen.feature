Feature: Authentication when user get total quiz of los
    Background:
        Given a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 3 quizzes

    Scenario Outline: User get total quiz of los
        Given a signed in "<role>"
        When user get total quiz of lo "1" with role
        Then returns "<status>" status code
        Examples:
            | role         | status |
            | school admin | OK     |
            | hq staff     | OK     |
            | student      | OK     |
            | teacher      | OK     |
            | parent       | OK     |
# | center manager | PermissionDenied |
# | center staff   | PermissionDenied |
