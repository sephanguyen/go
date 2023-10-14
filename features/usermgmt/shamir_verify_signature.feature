Feature: Signature verification
  As a school staff
  I need to be able to create a new student with grade master

  Scenario Outline: Signature verification with invalid request
    Given a signed in "school admin"
    And system run job to generate API Key with organization "MANABIE_SCHOOL" 
    And an invalid VerifySignatureRequest with "<condition>"
    When a client verifies signature
    Then receives "PermissionDenied" status code

    Examples:
      | condition           |
      | invalid signature   |
      | invalid public key  |
      | api key was deleted |

  Scenario Outline: Signature verification with valid request
    Given a signed in "school admin"
    And system run job to generate API Key with organization "MANABIE_SCHOOL" 
    And a valid VerifySignatureRequest
    When a client verifies signature
    Then receives "OK" status code