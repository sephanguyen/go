Feature: Finish brightcove upload url v2
  Background: create birghtcove upload url
    Given a signed in "school admin"
    When api v2 user create brightcove upload url for video "manabie.mp4"
    Then returns "OK" status code
    And api v2 yasuo must return a video upload url

  Scenario Outline: user with invalid role try to finish brightcove upload url
    Given a signed in "<signed as>"
    When api v2 user finish brightcove upload url for video "manabie.mp4"
    Then returns "<msg>" status code

    Examples:
      | signed as       | msg              |
      | unauthenticated | Unauthenticated  |
      | student         | PermissionDenied |
      | parent          | PermissionDenied |

  Scenario: school admin create brightcove upload url
    Given a signed in "school admin"
    When api v2 user finish brightcove upload url for video ""
    Then returns "InvalidArgument" status code

  Scenario: school admin create brightcove upload url
    Given a signed in "school admin"
    When api v2 user finish brightcove upload url for video "manabie.mp4"
    Then returns "OK" status code

