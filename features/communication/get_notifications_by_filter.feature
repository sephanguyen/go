Feature: user gets notification by filters

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin has created some tags

    Scenario Outline: staff get notifications by empty filter with granted role
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "10" draft and "5" schedule after one day to some students
        And sends "4" of drafted notifications
        And discards "3" of drafted notifications
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "12" notifications in CMS with corrected data
        And see "3" drafted and "4" sent and "5" scheduled and "12" total notifications count in CMS notification tab
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario Outline: staff get notifications by empty filter with granted denied role
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "10" draft and "5" schedule after one day to some students
        And sends "4" of drafted notifications
        And discards "3" of drafted notifications
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    @blocker
    Scenario Outline: staff get notifications by empty filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        And discards "<num discard>" of drafted notifications
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num notification>" notifications in CMS with corrected data
        And see "<num draft remaining>" drafted and "<num send>" sent and "<num schedule>" scheduled and "<num notification>" total notifications count in CMS notification tab
        Examples:
            | num draft | num schedule | num send | num discard | num draft remaining | num notification |
            | 10        | 5            | 4        | 3           | 3                   | 12               |
            | 13        | 2            | 1        | 0           | 12                  | 15               |

    Scenario Outline: staff get notifications by tag filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        And attach some tags for "<num notitag>" notifications
        When current staff get notifications by filter with status is "all" and tags is "all tags added" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num notitag>" notifications in CMS with corrected data
        And see "auto detect" drafted and "auto detect" sent and "auto detect" scheduled and "<num notitag>" total notifications count in CMS notification tab
        Examples:
            | num draft | num schedule | num send | num notitag |
            | 10        | 5            | 4        | 12          |
            | 13        | 2            | 1        | 2           |

    Scenario Outline: staff get notifications by tag filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        And attach some tags for "<num notitag>" notifications
        And remove all tags on all notifications
        When current staff get notifications by filter with status is "all" and tags is "all tags added" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num notitag>" notifications in CMS with corrected data
        Examples:
            | num draft | num schedule | num send | num notitag |
            | 10        | 5            | 4        | 0           |

    Scenario Outline: staff get notifications by status filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        When current staff get notifications by filter with status is "<status>" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num notistatus>" notifications in CMS with corrected data
        And see "<num draft remaining>" drafted and "<num send>" sent and "<num schedule>" scheduled and "<total count>" total notifications count in CMS notification tab
        Examples:
            | num draft | num schedule | num send | num draft remaining | num notistatus | total count | status    |
            | 10        | 5            | 4        | 6                   | 4              | 15          | sent      |
            | 3         | 2            | 1        | 2                   | 2              | 5           | scheduled |
            | 8         | 2            | 2        | 6                   | 6              | 10          | draft     |

    Scenario Outline: staff get notifications by send time filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "<send from>" and send time to is "<send to>" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num send result>" notifications in CMS with corrected data
        And see "0" drafted and "<num send result>" sent and "0" scheduled and "<num send result>" total notifications count in CMS notification tab
        Examples:
            | num draft | num schedule | num send | num send result | send from    | send to      |
            | 10        | 5            | 4        | 4               | 1 min before | none         |
            | 13        | 2            | 1        | 1               | none         | 1 min after  |
            | 4         | 2            | 1        | 0               | 1 min after  | none         |
            | 4         | 1            | 2        | 0               | none         | 1 min before |
            | 2         | 2            | 2        | 2               | 1 min before | 1 min after  |

    Scenario Outline: staff get notifications by title filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert "<num notifciations>" draft notifications with "<num have title>" have title is "notification title" and the rest have random title
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "notification title" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num have title>" notifications in CMS with corrected data
        And see "<num have title>" drafted and "0" sent and "0" scheduled and "<num have title>" total notifications count in CMS notification tab
        Examples:
            | num notifciations | num have title |
            | 13                | 2              |
            | 4                 | 2              |

    Scenario: staff get notifications with paging filter
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        When current staff get notifications by filter with status is "<status>" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "<limit>" and offset is "<offset>" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num notification in filtered>" notifications in CMS with corrected data
        And see "<num draft>" drafted and "0" sent and "<num schedule>" scheduled and "<num total notification>" total notifications count in CMS notification tab
        And see previous offset is "<previous offset>" and next offset is "<next offset>"
        Examples:
            | num draft | num schedule | offset | limit | num notification in filtered | previous offset | next offset | num total notification |
            | 7         | 1            | 0      | 1     | 1                            | 0               | 1           | 8                      |
            | 9         | 2            | 0      | 5     | 5                            | 0               | 5           | 11                     |
            | 8         | 0            | 6      | 3     | 2                            | 3               | 8           | 8                      |
            | 6         | 6            | 3      | 10    | 9                            | 0               | 12          | 12                     |

    Scenario: filter notification by target group
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert notification to "student, parent" and "<course_filter>" course and "all" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "<target group filter>" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "<num noti>" notifications in CMS with corrected data
        Examples:
            | course_filter | location_filter | class_filter | school_filter | target group filter                    | num noti |
            | all           | all             | all          | all           | all location                           | 1        |
            | all           | all             | all          | all           | all course                             | 1        |
            | all           | all             | all          | all           | all class                              | 1        |
            | random        | random          | random       | random        | list location                          | 1        |
            | random        | random          | random       | random        | list course                            | 1        |
            | random        | random          | random       | random        | list class                             | 1        |
            | random        | random          | random       | random        | list location, list course, list class | 1        |

    Scenario: Get fully questionnaire submitted
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "2" students with the same parent
        And school admin creates "1" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        And a questionnaire with resubmit allowed "false", questions "1.multiple_choice, 2.free_text.required, 3.check_box.required" respectively
        When current staff send a notification with attached questionnaire to "<receivers>"
        Then parent answer questionnaire for "<answered for students>"
        When current staff get notifications by filter with status is "all" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "none"
        Then returns "OK" status code
        And see "1" notifications in CMS with corrected data
        Examples:
            | receivers | answered for students |
            | parent    | 1,2                   |

    Scenario: filter notification by composer
        Given school admin creates "random" students
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert some notifications with "<num draft>" draft and "<num schedule>" schedule after one day to some students
        And sends "<num send>" of drafted notifications
        When current staff get notifications by filter with status is "<status>" and tags is "none" and send time from is "none" and send time to is "none" and title is "none" and target_group filter is "none" and limit is "none" and offset is "none" and fully questionnaire submitted is "false" and composer is "current"
        Then returns "OK" status code
        And see "<num noti>" notifications in CMS with corrected data
        Examples:
            | num draft | num schedule | num send | num noti |
            | 10        | 5            | 4        | 15       |
            | 3         | 2            | 1        | 5        |
