@quarantined
Feature: Sync Student Subscription

  Scenario Outline: Sync Student Subscription successfully
    # Test `BULKUPSERT` from eureka.course_students then sync to bob.lesson_student_subscriptions
    # Gen Students & Courses
    Given a "<number-of-students>" number of existing students
    And a "<number-of-courses>" number of existing courses
    And a list of locations with are existed in DB
    # Upsert & Sync student subscription
    And a signed in as "<signed-in-user>"
    And returns "OK" status code
    When assigning course packages to existing students
    Then returns "OK" status code
    And sync student subscription successfully

    Examples:
      | signed-in-user | number-of-students | number-of-courses |
      | school admin   | 10                 | 10                |
      | school admin   | 20                 | 10                |
      | school admin   | 20                 | 1                 |

  @quarantined
  Scenario Outline: Sync Updated Student Subscription successfully
    # Test `UPDATE` from eureka.course_students then sync to bob.lesson_student_subscriptions
    Given an existing student
    And an existing course
    And a list of locations with are existed in DB
    And a signed in as "<signed-in-user>"
    And returns "OK" status code
    When assigning course packages to existing students
    And returns "OK" status code
    When edit course package with new start at "<start-at>" and end at "<end-at>"
    And returns "OK" status code
    Then sync student subscription with new start at "<start-at>" and end at "<end-at>" successfully

    Examples:
      | signed-in-user | start-at                 | end-at                   |
      | admin          | 2010-06-30T23:59:59.000Z | 2060-06-30T23:59:59.000Z |
