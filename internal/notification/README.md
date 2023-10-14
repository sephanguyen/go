### Notification service
Code structure:
```
└── internal
    └── notification
        ├── common
        ├── config
        ├── infra
        │   ├── events.go
        │   ├── firebase.go
        │   ├── logging.go
        │   └── metrics.go
        ├── model
        │   ├── cources.go
        │   ├── notifications.go
        │   └── user.go
        ├── repository
        │   ├── notifications.go
        │   └── notifications_test.go
        ├── service
        │   ├── notification_service.go
        │   └── validation
        │       └── notification.go
        └── transport
        │   ├── grpc
        │   │   ├── notification_modifier.go
        │   │   └── notification_reader.go
        │   └── nats
        │       ├── cources_sync.go
        │       ├── notification_requests.go
        │       └── users_sync.go
        └── README.md
```

