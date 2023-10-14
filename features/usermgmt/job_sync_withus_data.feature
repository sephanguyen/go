@blocker
Feature: Run Job sync withus data
  Manabie system need to sync withus data from tsv file in google storage bucket

  Scenario Outline: Sync data successfully
    Given schedule job uploaded a tsv data file in "<org>" bucket
    When system run job to sync data from bucket
    Then data were synced successfully

    Examples:
      | org    |
      | itee   |
      | withus |

  # Scenario Outline: Sync data failed when object does not exist
  #   Given system does not have tsv data file in "<org>" bucket
  #   When system run job to sync data from bucket
  #   Then system sync data failed
  #   And previous data must not be changed

  #   Examples:
  #     | org    |
  #     | itee   |
  #     | withus |

  # Scenario Outline: Sync data failed when tsv file has invalid data
  #   Given tsv data file in "<org>" bucket which has invalid "<invalid data type>" data
  #   When system run job to sync data from bucket
  #   Then system sync data failed
  #   And previous data must not be changed

  #   Examples:
  #     | org    | invalid data type      |
  #     | itee   | tag                    |
  #     | withus | course                 |
  #     | itee   | location               |
  #     | withus | existed student number |
  #   # both username and email are same behavior
  #     | itee   | existed email          |
  #     | withus | existed email          |
