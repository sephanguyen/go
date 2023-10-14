# This test doesn't relate to domain features, its purposes are demonstrating and testing
Feature: Test unleash manager

  @blocker
    @feature_flag:Test_User_Feature_A:enable
    @feature_flag:Test_User_Feature_B:enable
  Scenario Outline:
    When a scenario requires "<feature flag names>" with corresponding statuses: "<feature flag statuses>"
    Then "<feature flag names>" must be locked and have corresponding statuses: "<feature flag statuses>"

    Examples:
      | feature flag names                      | feature flag statuses |
      | Test_User_Feature_A,Test_User_Feature_B | enable,enable         |

  @blocker
    @feature_flag:Test_User_Feature_B:disable
    @feature_flag:Test_User_Feature_A:disable
  Scenario Outline:
    When a scenario requires "<feature flag names>" with corresponding statuses: "<feature flag statuses>"
    Then "<feature flag names>" must be locked and have corresponding statuses: "<feature flag statuses>"

    Examples:
      | feature flag names                      | feature flag statuses |
      | Test_User_Feature_B,Test_User_Feature_A | disable,disable       |
