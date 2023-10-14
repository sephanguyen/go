@quarantined
Feature: Get appsmith page info

  Scenario: Get appsmith page info
    Given "school admin" signin system
    When user gets appsmith page info by slug
    # Becauset the appsmith instance not allow run on locally, so temporarily skip this step: https://github.com/manabie-com/backend/blob/develop/skaffold.appsmith.yaml#L5
    # Then returns corresponding appsmith page
