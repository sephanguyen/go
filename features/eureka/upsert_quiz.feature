Feature: Upsert quiz

  Background:
    Given a signed in "school admin"

  Scenario Outline: Validate upsert chapters
    When user upsert "<validity>" quiz
    Then returns "<status>" status code

    Examples:
      | validity            | status          |
      | valid               | OK              |
      | missing ExternalId  | InvalidArgument |
      | missing Question    | InvalidArgument |
      | missing Explanation | InvalidArgument |
      | missing Options     | InvalidArgument |

  Scenario Outline: Upsert valid quiz
    Given a signed in "school admin"
    When user upsert a "<quiz>" quiz
    Then returns "OK" status code
    Then quiz created successfully with new version
    And quiz_set also updated with new version

    Examples:
      | quiz              |
      | valid             |
      | existed           |
      | emptyTagLO        |
      | termAndDefinition |
      | fillInTheBlank    |
