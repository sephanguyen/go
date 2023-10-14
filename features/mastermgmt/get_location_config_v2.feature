Feature: Get location configuration based on locations
    Scenario Outline: Get all configurations
        Given "school admin" signin system
        Given location configurations v2 value "<location_configs>" existed on DB
        When user gets locations configurations with "<list_locations>" locations
        Then returns "<status>" status code
        And locations configurations are returned with "<return_location_configs>"

        Examples:
            | location_configs                         | list_locations | status | return_location_configs                   |
            | loc1:enabled, loc2:enabled, loc3:enabled | loc1,loc2,loc3 | OK     | loc1:enabled, loc2:enabled, loc3:enabled  |
            | loc1:enabled, loc2:disable, loc3:enabled | loc1,loc2,loc3 | OK     | loc1:enabled, loc2:disabled, loc3:enabled |
            | loc1:enabled, loc2:disable, loc3:enabled | loc1,loc2      | OK     | loc1:enabled, loc2:disabled               |
