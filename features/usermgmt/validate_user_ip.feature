@blocker
Feature: validation check for user IP address

    Scenario Outline: Validate user IP address
        Given a signed in "staff granted role school admin"
        And school admin's IP is "<ipType>" whitelist and the IP restriction feature is "<featureAction>"
        When school admin validates the IP address
        Then school admin sees the IP address is "<permission>"
        Examples:
            | ipType | featureAction | permission  |
            | in     | on            | allowed     |
            | not in | on            | not allowed |
            | not in | off           | allowed     |

