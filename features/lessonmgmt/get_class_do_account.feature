Feature: Get ClassDo Account

    Scenario: Get ClassDo Account
        Given user signed in as school admin
        And have some imported ClassDo accounts
        When user gets Class Do by User ID
        Then returns "OK" status code
        And got expected ClassDo account
