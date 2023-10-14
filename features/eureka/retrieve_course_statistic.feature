@quarantined
Feature: Retrieve course statistic

  Background: Valid content books studyplan for that books
    Given "school admin" logins "CMS"
    And "teacher" logins "Teacher App"

  Scenario Outline: retrieve course statistic
    Given 2 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 2 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And 2 students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When teacher retrieve course statistic
    And our system returns correct course statistic

  Examples:
    | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics |
    | 2        | 3               | 1                | 5               |  1              |
    | 3        | 3               | 2                | 10              |  0              |
    | 4        | 3               | 4                | 10              |  0              |

  Scenario Outline: some study plan items are archived
    Given 2 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 2 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And <num_student> students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    And some of created study plan item are archived
    When teacher retrieve course statistic
    And our system returns correct course statistic

  Examples:
    |   num_student   | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics |
    |       1         | 2        | 3               | 1                | 5               |  1              |
    |       2         | 3        | 3               | 2                | 10              |  0              |
    |       2         | 4        | 3               | 4                | 10              |  0              |
    |       0         | 0        | 0               | 0                | 0               |  0              |


Scenario Outline: retrieve course statistic with class filter
    Given 2 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 2 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And some students are members of some classes
    And 2 students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When teacher retrieve course statistic
    And our system returns correct course statistic

  Examples:
    | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics |
    | 2        | 3               | 1                | 5               |  1              |
    | 3        | 3               | 2                | 10              |  0              |
    | 4        | 3               | 4                | 10              |  0              |

Scenario Outline: correct score for 1 student
    Given 1 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 1 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And <num_student> students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When teacher retrieve course statistic
    And our system returns correct course statistic
    And topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>

  Examples:
    | num_student | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics | topic_assigned | topic_completed_student | topic_average_score |
    | 1           | 2        | 5               | 2                | 7               |  0              |    1           |        1                |    85               |
    | 1           | 1        | 4               | 1                | 3               |  0              |    1           |        0                |    55               |
    | 1           | 1        | 3               | 2                | 0               |  0              |    1           |        0                |    20               |
    | 1           | 2        | 0               | 2                | 0               |  0              |    1           |        1                |    0                |
    | 1           | 0        | 0               | 0                | 0               |  0              |    1           |        0                |    0                |

Scenario Outline: correct score for all student
    Given 2 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 1 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And <num_student> students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When teacher retrieve course statistic
    And our system returns correct course statistic
    And topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>

  Examples:
    | num_student | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics | topic_assigned | topic_completed_student | topic_average_score |
    | 2           | 4        | 5               | 4                | 7               |  0              |    2           |        2                |    85               |
    | 2           | 4        | 2               | 4                | 3               |  0              |    2           |        2                |    35               |
    | 2           | 4        | 0               | 4                | 0               |  0              |    2           |        2                |    0                |
    | 2           | 0        | 0               | 0                | 0               |  0              |    2           |        0                |    0                |

Scenario Outline: correct score for all student v2
    Given 2 students logins "Learner App"
    And "school admin" has created a book with each 2 los, 2 assignments, 1 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for all login student
    And <num_student> students do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When <course_statistical_v2>teacher retrieve course statistic
    And <course_statistical_v2>our system returns correct course statistic
    And <course_statistical_v2>topic total assigned student is <topic_assigned>, completed students is <topic_completed_student>, average score is <topic_average_score>

  Examples:
    | num_student | done_los | correct_quizzes | done_assignments | assignment_mark |  skipped_topics | topic_assigned | topic_completed_student | topic_average_score |
    | 2           | 4        | 5               | 4                | 7               |  0              |    2           |        2                |    85               |
    | 2           | 4        | 2               | 4                | 3               |  0              |    2           |        2                |    35               |
    | 2           | 4        | 0               | 4                | 0               |  0              |    2           |        2                |    0                |
    | 2           | 0        | 0               | 0                | 0               |  0              |    2           |        0                |    0                |

