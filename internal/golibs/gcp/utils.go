package gcp

//Temporarily disable for now
/*func newGCPAuthClientFromCredentialFile(ctx context.Context, credentialsFile string) (*auth.Client, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Println("error initializing app:", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Println("error getting Auth client:", err)
	}

	return authClient, err
}

func newGCPAuthClientFromCredentialJSON(ctx context.Context, credentialsJSON []byte) (*auth.Client, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		log.Println("error initializing app:", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Println("error getting Auth client:", err)
	}

	return authClient, err
}*/
