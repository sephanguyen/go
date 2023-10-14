package constant

import "time"

type Version string
type FileStoreName string

const (
	TouchEntry = "TOUCH_ENTRY"
	TouchExit  = "TOUCH_EXIT"

	TouchIntervalInMinutes = 1

	UserGroupStudent                 = "USER_GROUP_STUDENT"
	UserGroupAdmin                   = "USER_GROUP_ADMIN"
	UserGroupSchoolAdmin             = "USER_GROUP_SCHOOL_ADMIN"
	UserGroupParent                  = "USER_GROUP_PARENT"
	ClientIDNatsEntryexitmgmtService = "entry_exit_notify_client_id"

	ContentType = "image/png"
	PageLimit   = 100

	DayDuration = 24 * time.Hour

	PermissionDeniedErrMsg = "you don't have permission"

	V1 Version = ""
	V2 Version = "v2"

	GoogleCloudStorageService FileStoreName = "GCS"
	MinIOService              FileStoreName = "MinIO"

	PNG string = "png"

	PgConnDuplicateError  = "pgconn: duplicate error"
	PgConnForeignKeyError = "pgconn: foreign key error"
	StudentQrRLSError     = "new row violates row-level security policy for table \"student_qr\" (SQLSTATE 42501)"
	EntryExitQueueAbort   = "Cannot insert or update. Another record with id"

	RoleSchoolAdmin = "School Admin"
	RoleTeacher     = "Teacher"
	RoleParent      = "Parent"
	RoleStudent     = "Student"

	SynersiaResourcePath = "-2147483646"

	AutoGenQRCodeConfigKey = "entryexit.entryexitmgmt.enable_auto_gen_qrcode"
)
