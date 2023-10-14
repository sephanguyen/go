Feature: Tag student for sibling discount

  Scenario: Auto tag student for sibling discount
    Given prepare data for sibling discount automation for "<case>"
    When service sent order info stream
    Then student is "<event>" for sibling discount automation
    
    Examples:
      | case                                                | event               |
      | without sibling                                     | not tracked         |
      | not valid for tracking                              | not tracked         |
      | valid for tracking but not valid for tagging        | tracked             |
      | valid for tagging                                   | tagged              |