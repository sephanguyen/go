Feature: upsert staff transpostation expense config
  
  Scenario Outline: School Admin insert transportation expense config value for staff
    Given "staff granted role school admin" signin system
    And new insert staff transportation expense request with "<record number>"
    And remove all staff old staff transportation expense records
    When user upsert staff transportation expense config
    Then returns "<resp status-code>" status code

    Examples:
        | record number                                         | resp status-code  |
        | 5                                                     | OK                |


  Scenario Outline: School Admin update transportation expense config value for staff
    Given "staff granted role school admin" signin system
    And new update staff transportation expense config request
    And remove all staff old staff transportation expense records
    When user upsert staff transportation expense config
    Then returns "<resp status-code>" status code

    Examples:
        | resp status-code  |
        | OK                |

  
  Scenario Outline: School Admin insert and update transportation expense config value for staff
    Given "staff granted role school admin" signin system
    And new upsert staff transportation expense config request
    And remove all staff old staff transportation expense records
    When user upsert staff transportation expense config
    Then returns "<resp status-code>" status code

    Examples:
        | resp status-code  |
        | OK                |


  Scenario Outline: School Admin delete transportation expense config value for staff
    Given "staff granted role school admin" signin system
    And new delete staff transportation expense config request
    When user upsert staff transportation expense config
    Then returns "<resp status-code>" status code

    Examples:
        | resp status-code  |
        | OK                |