package entities

import "github.com/jackc/pgtype"

type File struct {
	FileID       pgtype.Text
	FileName     pgtype.Text
	FileType     pgtype.Text
	DownloadLink pgtype.Text
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (f *File) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"file_id",
		"file_name",
		"file_type",
		"download_link",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&f.FileID,
		&f.FileName,
		&f.FileType,
		&f.DownloadLink,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.DeletedAt,
		&f.ResourcePath,
	}
	return
}

func (*File) TableName() string {
	return "file"
}
