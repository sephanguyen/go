Feature: Manually Record Student Entry and Exit time
  As a school staff
  I am able to create a student entry and exit record
  Background:
    Given there is an existing student

  @major
  Scenario Outline: School staff creates entry and exit record successfully
    Given student has "Existing" parent
    And student parent has existing device
    When "<signed-in user>" "<checked-unchecked>" notify parents checkbox
    And "<signed-in user>" creates "<entry-exit>" record of this student in "<time-zone>"
    Then new entry exit record is created successfully
    And receives "OK" status code
    And parent receives notification status "<notif-status>"

    Examples:
      | signed-in user | entry-exit     | checked-unchecked | notif-status   | time-zone        |
      | school admin   | entry          | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | school admin   | entry and exit | checked           | Successfully   | Asia/Tokyo       |
      | hq staff       | entry          | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | hq staff       | entry and exit | checked           | Successfully   | Asia/Tokyo       |
      | centre lead    | entry          | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre manager | entry          | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre manager | entry and exit | checked           | Successfully   | Asia/Tokyo       |
      | centre staff   | entry          | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre staff   | entry and exit | checked           | Successfully   | Asia/Tokyo       |

  @major
  Scenario Outline: School staff creates invalid entry exit record
    Given student has "Existing" parent
    And student parent has existing device
    When "<signed-in user>" "checked" notify parents checkbox
    And "<signed-in user>" creates invalid "<invalid argument>" request
    Then receives "InvalidArgument" status code
    And parent receives notification status "Unsuccessfully"

    Examples:
      | signed-in user | invalid argument                       |
      | school admin   | no entry date                          |
      | school admin   | entry date is ahead than exit date     |
      | school admin   | entry time is ahead than exit time     |
      | school admin   | entry time is ahead than current time  |
      | school admin   | entry date is ahead than current date  |
      | school admin   | exit time is ahead than current time   |
      | hq staff       | no entry date                          |
      | hq staff       | entry date is ahead than exit date     |
      | centre lead    | entry time is ahead than exit time     |
      | centre lead    | entry time is ahead than current time  |
      | centre manager | entry date is ahead than current date  |
      | centre manager | exit time is ahead than current time   |
      | centre staff   | exit date is ahead than current date   |
      | centre staff   | cannot retrieve student id in database |

  @major
  Scenario Outline: School staff creates entry and exit record successfully with no student parent
    Given student has "No" parent
    When "<signed-in user>" "checked" notify parents checkbox
    And "<signed-in user>" creates "<entry-exit>" record of this student in "<time-zone>"
    Then new entry exit record is created successfully
    And receives "OK" status code
    And parent receives notification status "Unsuccessfully"

    Examples:
      | signed-in user | entry-exit     | time-zone        |
      | school admin   | entry          | Asia/Ho_Chi_Minh |
      | school admin   | entry and exit | Asia/Tokyo       |
      | hq staff       | entry          | Asia/Ho_Chi_Minh |
      | hq staff       | entry and exit | Asia/Tokyo       |
      | centre lead    | entry          | Asia/Ho_Chi_Minh |
      | centre lead    | entry and exit | Asia/Tokyo       |
      | centre manager | entry          | Asia/Ho_Chi_Minh |
      | centre manager | entry and exit | Asia/Tokyo       |
      | centre staff   | entry          | Asia/Ho_Chi_Minh |
      | centre staff   | entry and exit | Asia/Tokyo       |
