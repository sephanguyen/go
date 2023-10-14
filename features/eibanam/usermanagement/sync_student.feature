@runsequence
Feature: Checking login after sync data from partner

    Background:
        Given staff creates school admin account for partner manually
        And system has synced teacher from partner
        And "school admin" logins CMS
        And "teacher" logins Teacher App
        And system has synced course and 2 classes from partner

    # E2E tracking available class, update visible course to available course
    # Teacher see 2 place: to-review, study plan
    Scenario: View student list after sync data
        When system syncs student account which associate with all available course-class
        And school admin creates study plan for course
        Then student logins Learner App successfully with student partner account
        And student sees course which student joins on Learner App
        And school admin sees this student on student-study plan page
        And teacher sees this student info on Teacher App

    # Scenario: Unable to login juku student account after sync data for JPREP
    #     When system syncs juku student account which associate with all available class
    #     And student logins juku student account on Learner App
    #     Then student logins Learner App failed

    Scenario: Edit student name by sync data
        Given system has synced student account which associate with all available class
        When system syncs existed student with new name
        Then student sees the edited name on Learner App
        And teacher sees the edited name on Teacher App
        And school admin sees the edited name on student-study plan page

    # Scenario Outline: Edit class's duration from <oldStatus> to <newStatus> by sync student data
    #     Given system has synced student which associate with all "<oldStatus>" class
    #     When system syncs student with "<newStatus>" class's duration
    #     Then school admin sees course on CMS
    #     And teacher sees course on Teacher App
    #     And student "<result>" course on Learner App

    #     Examples:
    #         | oldStatus   | newStatus   | result       |
    #         | available   | unavailable | does not see |
    #         | unavailable | available   | sees         |

# Need to check
#    Scenario: Delete student by sync data
#        Given system has synced student which associate with all available class
#        When system syncs student with deleted student
#        Then student does not login successfully on Learner App
# And school admin does not see this student on Teacher App
# And teacher does not see this student on Teacher App
# For these case, I track on this ticket: teacher is show student on lesson/to-review