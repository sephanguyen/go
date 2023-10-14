@major
Feature: Scheduled Invoice Checker

	# Because of how InvoiceScheduleChecker run and process scheduled invoice, each scenarios should have different organizations and invoice date
	# to prevent flaky tests

	# To prevent conflicts with other tests, the organizations used in this test should not be used in other test and vice versa.

	# There should be specific organizations or resource_path that can and cannot have DRAFT invoices.
	# In this test, the below organizations will be having a generated scheduled invoice:
	# * -2147483635
	# * -2147483644

	Scenario Outline: All of the student with bill item successfully generated invoice from different org that have scheduled invoice
		Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And a bill item of these organization "<orgs-with-scheduled-invoice>" has adjustment price "4"
		And there is scheduled invoice to be run at day "0" for these organizations "<orgs-with-scheduled-invoice>"
		And "one" bill item has review required tag
		And "one" bill item created after the cutoff date
		When the InvoiceScheduleChecker endpoint was called at day "0"
		Then receives "OK" status code
		And there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And each invoice has correct total amount and outstanding balance
		And there are no students invoice generated in organizations "<orgs-without-scheduled-invoice>"
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved
		And only student billed bill items are invoiced
		And all bill item with review required tag was skipped
		And all bill item created after the cutoff date was skipped

		Examples:
			| orgs-with-scheduled-invoice | orgs-without-scheduled-invoice        |
			| -2147483635                 | -2147483643, -2147483628, -2147483641 |

	Scenario Outline: The invoice schedule was run concurrently
		Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And there is scheduled invoice to be run at day "3" for these organizations "<orgs-with-scheduled-invoice>"
		And "no" bill item has review required tag
		When the InvoiceScheduleChecker endpoint was called at day "3" concurrently
		Then only one response has OK status code and others have error
		And there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And each invoice has correct total amount and outstanding balance
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved
		And only student billed bill items are invoiced

		Examples:
			| orgs-with-scheduled-invoice |
			| -2147483626                 |

	Scenario Outline: There are no student with billing item to be invoiced on the scheduled invoice date
		Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<organizations>" have "1" students with "1" "<bill-item-status>" bill items
		And there is scheduled invoice to be run at day "<scheduled-day>" for these organizations "<organizations>"
		And "no" bill item has review required tag
		When the InvoiceScheduleChecker endpoint was called at day "<scheduled-day>"
		Then receives "OK" status code
		And there are no students invoice generated in organizations "<organizations>"
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved

		# Values of scheduled-day are either added or deducted from the day today
		# They have different values since the actual feature for cronjob will be run once in a day
		# Example: -1 will be yesterday
		Examples:
			| organizations | bill-item-status | scheduled-day |
			| -2147483645   | INVOICED         | -1            |
			| -2147483643   | PENDING          | 1             |
			| -2147483628   | CANCELLED        | 2             |
			| -2147483647   | WAITING_APPROVAL | -2            |

	Scenario: No scheduled invoice to be run
		Given unleash feature flag is "enable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<organizations>" have "1" students with "1" "BILLED" bill items
		And there is no scheduled invoice to be run at day "-3"
		And "no" bill item has review required tag
		When the InvoiceScheduleChecker endpoint was called at day "-3"
		Then receives "OK" status code
		And there are no students invoice generated in organizations "<organizations>"

		Examples:
			| organizations |
			| -2147483639   |

	# ----------------------------------- Payment_OrderManagement_BackOffice_Reviewed_Flag disabled --------------------------------

	@quarantined
	Scenario Outline: All of the student with bill item successfully generated invoice from different org that have scheduled invoice even if bill items have review required tag
		Given unleash feature flag is "disable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And a bill item of these organization "<orgs-with-scheduled-invoice>" has adjustment price "4"
		And there is scheduled invoice to be run at day "0" for these organizations "<orgs-with-scheduled-invoice>"
		And "all" bill item has review required tag
		And "one" bill item created after the cutoff date
		When the InvoiceScheduleChecker endpoint was called at day "0"
		Then receives "OK" status code
		And there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And each invoice has correct total amount and outstanding balance
		And there are no students invoice generated in organizations "<orgs-without-scheduled-invoice>"
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved
		And only student billed bill items are invoiced
		And all bill item created after the cutoff date was skipped

		Examples:
			| orgs-with-scheduled-invoice | orgs-without-scheduled-invoice        |
			| -2147483635                 | -2147483643, -2147483628, -2147483641 |

	@quarantined
	Scenario Outline: The invoice schedule was run concurrently even if bill items have review required tag
		Given unleash feature flag is "disable" with feature name "Payment_OrderManagement_BackOffice_Reviewed_Flag"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And there is scheduled invoice to be run at day "3" for these organizations "<orgs-with-scheduled-invoice>"
		And "all" bill item has review required tag
		When the InvoiceScheduleChecker endpoint was called at day "3" concurrently
		Then only one response has OK status code and others have error
		And there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And each invoice has correct total amount and outstanding balance
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved
		And only student billed bill items are invoiced

		Examples:
			| orgs-with-scheduled-invoice |
			| -2147483626                 |