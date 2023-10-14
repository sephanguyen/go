Feature: Update Student Profile
  In order to update my own profile
  As a student
  I need to perform profile update

  Scenario: registered student updates profile
    Given a signed in student
    And a valid updates profile request
    And his owned student UUID
    When user updates profile
    Then Bob must records student's profile update
    And Bob returns "OK" status code
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with message type *pb.UpdateProfileRequest

  Scenario: registered student updates school using existed school
    Given a signed in admin
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And a school name "S3", country "COUNTRY_VN", city "Hồ Chí Minh", district "Bình Thạnh"
    And admin inserts schools

    And a signed in student
    And a valid updates profile request
    And student selects school in country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And his owned student UUID
    When user updates profile
    Then Bob must records student's profile update
    And Bob returns "OK" status code
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with message type *pb.UpdateProfileRequest

  Scenario: registered student updates school by input new school
    Given a signed in admin
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And a school name "S3", country "COUNTRY_VN", city "Hồ Chí Minh", district "Bình Thạnh"
    And admin inserts schools

    And a signed in student
    And a valid updates profile request
    And student inputs new school in country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And his owned student UUID
    When user updates profile
    Then Bob must records student's profile update
    And Bob returns "OK" status code
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with message type *pb.UpdateProfileRequest


