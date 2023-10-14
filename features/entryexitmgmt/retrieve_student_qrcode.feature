@major
Feature: Retrieve Student QR Code
  As a student
  I am able to see my qr code display on learner app

  Background:
    Given there is an existing student

  Scenario Outline: Student displayed qr code successfully
    Given this student has "<qr-existence>" qr code record
    And student logins on Learner App
    When student is at the My QR Code screen
    And student requested qr code with "valid" payload
    Then student qr code is displayed "successfully"
    And receives "OK" status code

    Examples:
      | qr-existence |
      | Existing     |
      | Not Existing |

  Scenario Outline: Student with v1 version of qrcode updated to current version
    Given student has "v1" qr version
    And student logins on Learner App
    When student is at the My QR Code screen
    And student requested qr code with "valid" payload
    Then student qr code is displayed "successfully"
    And receives "OK" status code
    And student should have updated qrcode version

  Scenario Outline: Student with invalid payload displayed qr code unsuccessfully
    Given this student has "<qr-existence>" qr code record
    And student logins on Learner App
    When student is at the My QR Code screen
    And student requested qr code with "invalid" payload
    Then student qr code is displayed "unsuccessfully"
    And receives "InvalidArgument" status code

    Examples:
      | qr-existence |
      | Existing     |
      | Not Existing |
