@quarantined @disabled
Feature: migrate data for student_package and student_package_class from fatima
    Scenario Outline: migrate notification_student_course
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "10" students
        And school admin creates "5" courses
        And insert class for each course into database
        And insert student_package and student_package_class into fatima database
        When run MigrateStudentPackageAndPackageClass
        And waiting to sync process is finished
        Then synced data on bob database correctly
