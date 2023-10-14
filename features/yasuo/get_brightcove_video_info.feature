Feature: User gets Brightcove video info
    Scenario: User successfully gets info of a Brightcove video
        Given a signed in "school admin"
        When user gets info of a "valid" video
        Then returns "OK" status code
        And the correct info of the video is returned
    
    Scenario Outline: User cannot get info of an invalid Brightcove video
        Given a signed in "school admin"
        When user gets info of a "<invalid video type>" video
        Then returns "<code>" status code
        Examples:
            | invalid video type    | code              |
            | invalid               | Internal          |
            | not_playable          | PermissionDenied  |