Feature: Create Live Lesson Private Chat
  Background: lesson conversation background
	Given resource path of school "Manabie" is applied
  And a lesson conversation with "2" teachers and "1" students

  Scenario Outline: user create new private lesson conversation
    When "<user 1>" create new live lesson private conversation with "<user 2>"
    Then returns "OK" status code
    And "<user 1>" see the live lesson private conversation
    Examples:
    |user 1|user 2|
    |student|teacher|
    |teacher|student|
    |teacher|teacher|