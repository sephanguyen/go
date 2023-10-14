Feature: Finish brightcove upload url
    Background: create birghtcove upload url
        Given "staff granted role school admin" signin system
        When user create brightcove upload url for video "manabie.mp4"
        Then returns "OK" status code
        And bob must return a video upload url

    Scenario Outline: user with invalid role try to finish brightcove upload url
        Given "<signed as>" signin system
        When user finish brightcove upload url for video "manabie.mp4"
        Then returns "<msg>" status code

        Examples:
            | signed as       | msg              |
            | unauthenticated | Unauthenticated  |
            | parent          | PermissionDenied |

    Scenario: admin create brightcove upload url
        Given "staff granted role school admin" signin system
        When user finish brightcove upload url for video ""
        Then returns "InvalidArgument" status code

    Scenario Outline: admin create brightcove upload url
        Given "<signed as>" signin system
        When user finish brightcove upload url for video "manabie.mp4"
        Then returns "OK" status code

        Examples:
            | signed as                       |
            | staff granted role school admin |
            | student                         |
            | teacher                         |

