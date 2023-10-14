Feature: Migrate student_packages to student_package_access_path
  Background:
    Given some package data in db

  Scenario: Migrate student_packages to student_package_access_path
    Given a number of existing student packages
    When system run job to migrate student_packages to student_package_access_path
    Then student_packages and student_package_access_path are correspondent