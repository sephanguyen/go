package nats

const (
	SubjectNotificationCreated = "Notification.Created"
	StreamNotification         = "notification"
	DurableNotification        = "durable-notification"
	QueueNotification          = "queue-notification"
	DeliverNotification        = "deliver.notification"
)

var ClientIDsAccepted = []string{
	"entry_exit_notify_client_id",
	"syllabus_assigment_client_id",
	"virtual_classroom_client_id",
	"bdd_testing_client_id",
}
