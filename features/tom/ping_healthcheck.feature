Feature: Ping message to keep connection
    To make sure connection still alive
    User need to send message Ping every 5 seconds

	Background: user already subscribe to endpoint streaming event
		Given resource path of school "Manabie" is applied
        Given a valid user token

	 Scenario: user connected to subscribeV2 try to ping stream with custom endpoint to keep connection alive
		When user subscribe to endpoint subscribeV2
		Then user send ping subscribeV2 to stream via ping endpoint every 5 seconds
		And tom should "keep" this connection of subscribeV2 more than 15 seconds

	Scenario: user connected to subscribeV2 but not send ping message should be disconnected after 15 seconds
		When user subscribe to endpoint subscribeV2
		And tom should "disconnect" this connection of subscribeV2 more than 15 seconds

