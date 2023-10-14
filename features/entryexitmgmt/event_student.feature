@major
Feature: Event Student

  Scenario: Student account is created with qrcode
    Given a EvtUser with message "CreateStudent"
    And entryexitmgmt internal config "enable_auto_gen_qrcode" is "on"
    When yasuo send event EvtUser
    Then student must have qrcode

  Scenario: Student account is created with qrcode
    Given a EvtUser with message "CreateStudent"
    And entryexitmgmt internal config "enable_auto_gen_qrcode" is "off"
    When yasuo send event EvtUser
    Then student must have no qrcode
