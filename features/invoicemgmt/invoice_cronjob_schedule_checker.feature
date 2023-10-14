@major
Feature: Cronjob Scheduled Invoice Checker
	# -2147483648 organization used on this test just to prevent flaky test from other tests on creating scheduled invoice
	# Test data will run at day "0" which is today so that we can check Cronjob that will trigger today

	Scenario Outline: Students with bill item successfully generated invoice that have scheduled invoice today
		Given unleash feature flag is "disable" with feature name "BACKEND_Invoice_InvoiceManagement_ImproveImportInvoiceChecker"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And there is scheduled invoice to be run at day "0" for these organizations "<orgs-with-scheduled-invoice>"
		And "no" bill item has review required tag
		And "no" bill item created after the cutoff date
		When cronjob run InvoiceScheduleChecker endpoint today
		Then there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved

		Examples:
			| orgs-with-scheduled-invoice |
			| -2147483644                 |

	@quarantined
	Scenario Outline: Students with bill item successfully generated invoice that have scheduled invoice today
		Given unleash feature flag is "enable" with feature name "BACKEND_Invoice_InvoiceManagement_ImproveImportInvoiceChecker"
		And the organizations "<orgs-with-scheduled-invoice>" have "3" students with "5" "BILLED" bill items
		And there is scheduled invoice to be run at day "0" for these organizations "<orgs-with-scheduled-invoice>"
		And "no" bill item has review required tag
		And "no" bill item created after the cutoff date
		When cronjob run InvoiceScheduleChecker endpoint today
		Then there are correct number of students invoice generated in organizations "<orgs-with-scheduled-invoice>"
		And the scheduled invoice status is updated to "INVOICE_SCHEDULE_COMPLETED"
		And a history of scheduled invoice was saved
		And there are no invoice scheduled student was saved

		Examples:
			| orgs-with-scheduled-invoice |
			| -2147483644                 |
