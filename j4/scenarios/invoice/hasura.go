package invoice

import (
	"github.com/manabie-com/backend/j4/serviceutil"
)

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "Invoice_InvoicesV2",
			Query: Invoice_InvoicesV2,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"limit":  10,
					"offset": 0,
				}
			},
		},
	}
	Invoice_InvoicesV2 = `
          query Invoice_InvoicesV2($limit: Int = 10, $offset: Int = 0,
          $invoice_order_by: invoice_order_by! = {created_at: desc}) {
            invoice(limit: $limit, offset: $offset, order_by: [$invoice_order_by]) {
              ...InvoiceAttrs
              payments {
                ...PaymentAttrs
              }
            }
            invoice_aggregate {
              aggregate {
                count
              }
            }
          }


          fragment InvoiceAttrs on invoice {
            invoice_id
            invoice_sequence_number
            status
            student_id
            sub_total
            total
            type
            created_at
          }


          fragment PaymentAttrs on payment {
            payment_date
            payment_due_date
            payment_expiry_date
            payment_method
            payment_status
          }`
)
