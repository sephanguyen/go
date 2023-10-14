package entities

import "github.com/jackc/pgtype"

type BulkPaymentRequestFile struct {
	BulkPaymentRequestFileID   pgtype.Text
	BulkPaymentRequestID       pgtype.Text
	FileName                   pgtype.Text
	FileURL                    pgtype.Text `sql:"file_url"`
	FileSequenceNumber         pgtype.Int4
	TotalFileCount             pgtype.Int4
	ParentPaymentRequestFileID pgtype.Text
	IsDownloaded               pgtype.Bool
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	DeletedAt                  pgtype.Timestamptz
	ResourcePath               pgtype.Text
}

func (e *BulkPaymentRequestFile) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_request_file_id",
			"bulk_payment_request_id",
			"file_name",
			"file_url",
			"file_sequence_number",
			"is_downloaded",
			"total_file_count",
			"parent_payment_request_file_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BulkPaymentRequestFileID,
			&e.BulkPaymentRequestID,
			&e.FileName,
			&e.FileURL,
			&e.FileSequenceNumber,
			&e.IsDownloaded,
			&e.TotalFileCount,
			&e.ParentPaymentRequestFileID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BulkPaymentRequestFile) TableName() string {
	return "bulk_payment_request_file"
}
