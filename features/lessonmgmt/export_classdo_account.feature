Feature: Export ClassDo Account

    Scenario: Export ClassDo Account CSV
      Given user signed in as school admin
      And have some imported ClassDo accounts
      When user export ClassDo accounts
      Then returns "OK" status code
