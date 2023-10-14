Feature: Update Student Entry and Exit records
  As a school staff
  I am able to update a student entry and exit record
  Background:
    Given there is an existing student

  @major
  Scenario Outline: School staff updates entry and exit record successfully
    Given student has "Existing" parent
    And student parent has existing device
    And student has "<entry-exit>" record
    When "<signed-in user>" "<checked-unchecked>" notify parents checkbox
    And "<signed-in user>" updates the "<entry-exit>" record of this student in "<time-zone>"
    Then entry exit record is updated successfully
    And receives "OK" status code
    And parent receives notification status "<notif-status>"

    Examples:
      | signed-in user | entry-exit | checked-unchecked | notif-status   | time-zone        |
      | school admin   | entry      | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | school admin   | exit       | checked           | Successfully   | Asia/Tokyo       |
      | hq staff       | entry      | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | hq staff       | exit       | checked           | Successfully   | Asia/Tokyo       |
      | centre lead    | entry      | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre manager | entry      | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre manager | exit       | checked           | Successfully   | Asia/Tokyo       |
      | centre staff   | entry      | unchecked         | Unsuccessfully | Asia/Ho_Chi_Minh |
      | centre staff   | exit       | checked           | Successfully   | Asia/Tokyo       |

  @major
  Scenario Outline: School staff updates invalid entry exit record
    Given student has "Existing" parent
    And student parent has existing device
    And student has "<entry-exit>" record
    When "<signed-in user>" "checked" notify parents checkbox
    And "<signed-in user>" updates the "<entry-exit>" record with invalid "<invalid argument>" request
    Then receives "InvalidArgument" status code
    And parent receives notification status "Unsuccessfully"

    Examples:
      | signed-in user | invalid argument                          | entry-exit |
      | school admin   | cannot retrieve entry exit id in database | entry      |
      | school admin   | no entry date                             | exit       |
      | school admin   | entry date is ahead than exit date        | entry      |
      | school admin   | entry time is ahead than exit time        | exit       |
      | school admin   | entry time is ahead than current time     | entry      |
      | school admin   | entry date is ahead than current date     | entry      |
      | school admin   | exit time is ahead than current time      | exit       |
      | school admin   | exit date is ahead than current date      | exit       |
      | school admin   | cannot retrieve student id in database    | entry      |

  @major
  Scenario Outline: School staff updates entry and exit record successfully with no student parent
    Given student has "No" parent
    And student has "<entry-exit>" record
    When "<signed-in user>" "checked" notify parents checkbox
    And "<signed-in user>" updates the "<entry-exit>" record of this student in "<time-zone>"
    Then entry exit record is updated successfully
    And receives "OK" status code
    And parent receives notification status "Unsuccessfully"

    Examples:
      | signed-in user | entry-exit | time-zone        |
      | school admin   | entry      | Asia/Ho_Chi_Minh |
      | school admin   | exit       | Asia/Tokyo       |
      | hq staff       | entry      | Asia/Ho_Chi_Minh |
      | hq staff       | exit       | Asia/Tokyo       |
      | centre lead    | entry      | Asia/Ho_Chi_Minh |
      | centre lead    | exit       | Asia/Tokyo       |
      | centre manager | entry      | Asia/Ho_Chi_Minh |
      | centre manager | exit       | Asia/Tokyo       |
      | centre staff   | entry      | Asia/Ho_Chi_Minh |
      | centre staff   | exit       | Asia/Tokyo       |
