Feature: get presigned url to upload

 Scenario: unauthenticated user try to check a class
   Given an invalid authentication token
   When user get url to upload file
   Then returns "Unauthenticated" status code

 Scenario Outline: User get a presigned url to upload file
   #Scenario: user try get url that can be used to upload file
   Given a signed in user has a expiration time "<expiration time>" and a prefix name "<prefix name>"
   When user get url to upload file
   Then return a presigned url to upload file and a expiration time "<actual expiration time>"

   #Scenario: user use presigned url to upload file
   When user wait a interval "<interval>"
   And upload a file via a presigned url
   Then return a status code "<status code>"
   And file storage must store file if presigned url not yet expired

   Examples:
     | expiration time | prefix name | actual expiration time | interval | status code |
     | 10s             | car         | 10s                    | 0        | 2xx         |
     | 0s              | bike        | 60s                    | 0        | 2xx         |
     | 20s             |             | 20s                    | 1        | 2xx         |
     | -1s             | bus         | 60s                    | 0        | 2xx         |
     | 20s             | bus         | 20s                    | 0        | 2xx         |
     | 1s              | bicycle     | 1s                     | 2        | 4xx         |
