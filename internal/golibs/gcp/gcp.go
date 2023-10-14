package gcp

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

const GoogleAppCredEnv = "GOOGLE_APPLICATION_CREDENTIALS"

type App struct {
	*firebase.App

	config AppConfig

	credentialFile string //deprecate soon
	ProjectID      string //deprecate soon
	ProjectConfig  *ProjectConfig
}

func NewApp(ctx context.Context, credentialFile string, projectID string) (*App, error) {
	if credentialFile == "" {
		credentialFile = os.Getenv(GoogleAppCredEnv)
	}

	opts := []option.ClientOption{
		option.WithCredentialsFile(credentialFile),
	}

	firebaseConfig := &firebase.Config{
		ProjectID: projectID,
	}

	firebaseApp, err := firebase.NewApp(ctx, firebaseConfig, opts...)
	if err != nil {
		return nil, err
	}

	app := &App{
		App:            firebaseApp,
		ProjectID:      projectID,
		credentialFile: credentialFile,
	}

	app.ProjectConfig, err = app.GetProjectConfig(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
}

type AppConfig interface {
	GetGCPProjectID() string
	GetGCPServiceAccountID() string
}

func appConfigToFirebaseConfig(config AppConfig) *firebase.Config {
	firebaseConfig := &firebase.Config{}

	if projectID := config.GetGCPProjectID(); projectID != "" {
		firebaseConfig.ProjectID = projectID
	}

	if serviceAccountID := config.GetGCPServiceAccountID(); serviceAccountID != "" {
		firebaseConfig.ServiceAccountID = serviceAccountID
	}

	return firebaseConfig
}

func NewGCPApp(ctx context.Context, credentialFile string, config AppConfig) (*App, error) {
	var firebaseAppOptions []option.ClientOption

	if credentialFile != "" {
		firebaseAppOptions = append(firebaseAppOptions, option.WithCredentialsFile(credentialFile))
	}

	firebaseApp, err := firebase.NewApp(ctx, appConfigToFirebaseConfig(config), firebaseAppOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp")
	}

	app := &App{
		App:       firebaseApp,
		ProjectID: config.GetGCPProjectID(),
		config:    config,
	}

	projectConfig, err := app.GetProjectConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetProjectConfig")
	}
	app.ProjectConfig = projectConfig

	return app, nil
}
