@blocker
Feature: Import Student

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create students by import csv with "<row condition>" successfully
    When school admin create 2 students with "<row condition>" by import in folder "students"
    Then students were upserted successfully by import

    Examples:
      | row condition                |
      | mandatory fields             |
      | all fields                   |
      | external user id with spaces |

  Scenario Outline: Create students by import invalid csv file with "<row condition>" unsuccessfully
    When school admin create 1 students with "<condition>" by import in folder "students"
    Then student were created unsuccessfully by import with code "<code>" and field "<field>" at row "2"

    Examples:
      | condition             | field    | code  |
      | wrong format birthday | birthday | 40004 |
      | type text gender      | gender   | 40004 |
      | out of gender range   | gender   | 40004 |
      | non-existing location | location | 40400 |

  Scenario Outline: Create students school histories by import csv with "<row condition>" successfully
    When school admin create 1 students with "<row condition>" by import in folder "school_histories"
    Then students were upserted successfully by import

    Examples:
      | row condition                                                            |
      | only school                                                              |
      | school and school_course                                                 |
      | school and school_course and start_date                                  |
      | there is current_school                                                  |
      | there is not current_school                                              |
      | 3 school 2 school_course 1 empty school_course empty start_date end_date |
      | 2 school 2 school_course 1 start_date 1 end_date                         |

  Scenario Outline: Create students invalid school histories by import csv with "<row condition>" unsuccessfully
    When school admin create 1 students with "<row condition>" by import in folder "school_histories"
    Then student were created unsuccessfully by import with code "<code>" and field "<field>" at row "2"

    Examples:
      | row condition                                    | field         | code  |
      | empty school but there is other fields           | school_course | 40004 |
      | school_course is not mapped to school            | school_course | 40004 |
      | non existing school                              | school        | 40400 |
      | non existing school_course                       | school_course | 40400 |
      | start_date after end_date                        | start_date    | 40004 |
      | 2 school 2 school_course 2 start_date 3 end_date | school        | 40004 |
      | 2 school 2 school_course 3 start_date 2 end_date | school        | 40004 |
      | 2 school 3 school_course 2 start_date 2 end_date | school        | 40004 |

  Scenario Outline: Create students student_tag by import csv with "<row condition>" unsuccessfully
    When school admin create 1 students with "<row condition>" by import in folder "student_tags"
    Then students were upserted successfully by import

    Examples:
      | row condition                              |
      | single student_tag                         |
      | multiple student_tag                       |
      | single discount student_tag                |
      | both non discount and discount student_tag |

  Scenario Outline: Create students invalid student_tag by import csv with "<row condition>" unsuccessfully
    When school admin create 1 students with "<row condition>" by import in folder "student_tags"
    Then student were created unsuccessfully by import with code "<code>" and field "<field>" at row "2"

    Examples:
      | row condition                      | field       | code  |
      | non existing student_tag           | student_tag | 40400 |
      | student_tag is parent_tag          | student_tag | 40004 |
      | one student_tag and one parent_tag | student_tag | 40004 |