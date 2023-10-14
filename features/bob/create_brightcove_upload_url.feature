Feature: Create brightcove upload url

    Scenario Outline: user with invalid role try to create brightcove upload url
        Given "<signed-in user>" signin system
        When user create brightcove upload url for video "manabie.mp4"
        Then returns "<msg>" status code

        Examples:
            | signed-in user  | msg              |
            | unauthenticated | Unauthenticated  |
            | parent          | PermissionDenied |

    Scenario Outline: admin create brightcove upload url
        Given "<signed-in user>" signin system
        When user create brightcove upload url for video "manabie.mp4"
        Then returns "OK" status code
        And bob must return a video upload url

        Examples:
            | signed-in user |
            | student        |
            | teacher        |

