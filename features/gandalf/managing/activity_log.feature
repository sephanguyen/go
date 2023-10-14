@zeus
Feature: Write activitylog of Zeus

  Background:
    Given data in table activity log is empty

  Scenario: Call a large amount request to each services (Bob, Tom, Eureka, Fatima, Yasuo, Shamir)
    Given a signed in "admin" with school: 3
    And 5 request update user profile with user group: "USER_GROUP_ADMIN", name: "update-user-%d", phone: "+849%d", email: "update-user-%d@example.com", school: 3

    # And a school with random name background
    # And 5 request update school with country "COUNTRY_VN", city "Thành phố Hồ Chí Minh", district "1"

    And a valid course background
    And 5 request list class by course

    And 5 request create package

    And a lesson conversation background
    And 5 request get conversation using current conversation id

    When all of above request are sent
    Then number of record in table activity log is 20