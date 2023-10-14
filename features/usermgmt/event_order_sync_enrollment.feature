@blocker
Feature: Sync enrollment status history from payment

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Enrollment status history sync from payment and update enrollment status successfully by "<order_type>"
    Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
    When school admin "<order_type>" to the student
    Then the enrollment status histories of student must "be updated"

    Examples:
      | order_type                                                            |
      | submit graduate order to the existed location                         |
      | submit withdrawal order to the existed location                       |
      | submit enrolled order to the existed location                         |
      | submit enrolled order to the location has temporary enrollment status |
      | submit LOA order to the existed location                              |
      | submit resume order to the existed location                           |

  Scenario Outline: Create new order to new location for student with order flow
    Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
    When school admin "<order_type>" to the student
    Then the enrollment status histories of student must "be potential"

    Examples:
      | order_type                                                       |
      | submit new order to the new location                             |
      | submit new order to the location has temporary enrollment status |

  Scenario Outline: Upsert students with updated end-date of temporary enrollment status history
    Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
    When school admin "<order_type>" to the student
    Then the enrollment status histories of student must "be updated"
    When school admin update a student with "update end-date temporary enrollment status history"
    Then students were upserted successfully

    Examples:
      | order_type                                      |
      | submit graduate order to the existed location   |
      | submit withdrawal order to the existed location |
      | submit enrolled order to the existed location   |

  Scenario Outline: Void an order and update enrollment status histories of a student with <void_order_type>
    Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
    When school admin "<submit_order_type>" to the student
    Then the enrollment status histories of student must "be updated"
    When school admin "<void_order_type>" to the student
    Then the enrollment status histories of student must "be removed correspondingly"

    Examples:
      | submit_order_type                               | void_order_type                               |
      | submit graduate order to the existed location   | void graduated order to the existed location  |
      | submit withdrawal order to the existed location | void withdrawal order to the existed location |
      | submit enrolled order to the existed location   | void enrolled order to the existed location   |
      | submit LOA order to the existed location        | void LOA order to the existed location        |
      | submit resume order to the existed location     | void resume order to the existed location     |

  Scenario: Void an order has location with potential enrollment status
    Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
    And school admin "submit new order to the location has temporary enrollment status" to the student
    When school admin "void order to the location has temporary enrollment status" to the student
    Then the enrollment status histories of student must "be removed correspondingly"

  Scenario: Submit an new order with same location and start date of deleted enrollment status history
    Given school admin create a student with "enrollment status history" and "potential and withdrawal status" by GRPC
    And enrollment history with "withdrawn" status of student was deleted
    When school admin "submit new withdrawal order with same location and start date of deleted enrollment status history" to the student
    Then the enrollment status histories of student must "be updated"


# @shanenoi will refactor with below format
#  Scenario Outline: creates enrollment request
#    Given school admin "creates" a student with "<initStatus>" status of location "<initLocation>" by GRPC
#    When school admin "creates enrollment request" to location "<submittedLocation>"
#    Then school admin sees "<deactivatedStatus>" status at location "<initLocation>" is "deactivated"
#    And school admin sees "<addedStatus>" status at location "<submittedLocation>" is "added"
#    Examples:
#      | initStatus | initLocation | submittedLocation | deactivatedStatus | addedStatus |
#      | potential  | A            | A                 | potential         | enrolled    |
#      | potential  | A            | B                 | no                | enrolled    |
#  # handle more cases related to creates enrollment request
#
#
#  Scenario Outline: create graduate request
#    Given school admin "creates" a student with "<initStatus>" status of location "<initLocation>" by GRPC
#    When school admin "creates graduate order" to location "<submittedLocation>"
#    Then school admin sees "<deactivatedStatus>" status at location "<initLocation>" is "deactivated"
#    And school admin sees "<addedStatus>" status at location "<submittedLocation>" is "added"
#    Examples:
#      | initStatus | initLocation | submittedLocation | deactivatedStatus | addedStatus |
#      | potential  | A            | A                 | potential         | graduate    |
#  # handle more cases related to creates graduate request
#
#  Scenario Outline: create withdrawn request
#    Given school admin "creates" a student with "<initStatus>" status of location "<initLocation>" by GRPC
#    When school admin "creates withdrawn order" to location "<submittedLocation>"
#    Then school admin sees "<deactivatedStatus>" status at location "<initLocation>" is "deactivated"
#    And school admin sees "<addedStatus>" status at location "<submittedLocation>" is "added"
#    Examples:
#      | initStatus | initLocation | submittedLocation | deactivatedStatus | addedStatus |
#  # handle more cases related to creates withdrawn request
#
#  Scenario Outline: create order request
#    Given school admin "creates" a student with "<initStatus>" status of location "<initLocation>" by GRPC
#    When school admin "creates order" to location "<submittedLocation>"
#    Then school admin sees "<deactivatedStatus>" status at location "<initLocation>" is "deactivated"
#    And school admin sees "<addedStatus>" status at location "<submittedLocation>" is "added"
#    Examples:
#      | initStatus | initLocation | submittedLocation | deactivatedStatus | addedStatus |
#  # handle more cases related to creates withdrawn request
#
#  Scenario Outline: void enrollment request
#    Given school admin "has created" a student with "<initStatus>" status of location "<initLocation>" by GRPC
#    And school admin "has submitted enrollment request" to location "<submittedLocation>"
#    When school admin "voids enrollment request" to location "<submittedLocation>"
#    And school admin sees "<submittedStatus>" status at location "<submittedLocation>" is "remove"
#    And school admin sees "<initStatus>" status at location "<initLocation>" is "reactivated"
#    Examples:
#      | initStatus | initLocation | submittedLocation | submittedStatus |
#      | potential  | A            | A                 | enrolled        |
#  # handle more cases related to void enrollment request
#  #  handle more cases related to void order type .....

# Scenario Outline: validate order flow
#   Given school admin "has created" a student with "<initStatus>" status of location "A" by GRPC
#   And school admin "<initRequest>" to location "A"
#   When school admin "<submitRequest>" to location "A"
#   And school admin sees "<submittedStatus>" status at location "<submittedLocation>" is "not synced"
#   Examples:
#     | initStatus | initRequest                      | submitRequest                      | submittedStatus |
#     | potential  | has submitted enrollment request | submits another enrollment request | enrolled        |
#     | potential  | has submitted withdrawn request  | submits another withdrawn request  | withdrawn       |
#     | potential  | has submitted graduate request   | submits another graduate request   | graduate        |
#     | temporary  | has submitted order request      | submits another order request      | potential       |

