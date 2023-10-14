Feature: Multiple Student interact with viewing qr code on learner App

  Scenario: Only display student qr code within the organization
    Given "student S1" with resource path from "organization -2147483648"
    And "student S1" has existing qr code
    And "student S2" with resource path from "organization -2147483646"
    And "student S2" has existing qr code
    When "<signed-in-user>" logins on Learner App
    Then "<signed-in-user>" "<result>" see "<student>" qr code

    Examples:
      | signed-in-user| result | student    |
      | student S1    | can    | student S1 |
      | student S1    | cannot | student S2 |
      | student S2    | can    | student S2 |
      | student S2    | cannot | student S1 |