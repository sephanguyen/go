# Google Cloud Functions

https://cloud.google.com/functions/docs/writing/background#functions_background_parameters-go
https://cloud.google.com/functions/docs/calling/firebase-auth

## How to deploy function by cmd

```
gcloud functions deploy ClaimTokenOnCreate \
 --entry-point ClaimTokenOnCreate \
 --trigger-event providers/firebase.auth/eventTypes/user.create \
 --trigger-resource [project-id] \
 --runtime [run-time] \
 --build-env-vars-file [env-file] \
 --service-account [service-account]
```

Example: Deploy ClaimTokenOnCreate fn on dev-manabie-online

```
gcloud functions deploy ClaimTokenOnCreate \
 --entry-point ClaimTokenOnCreate \
 --trigger-event providers/firebase.auth/eventTypes/user.create \
 --trigger-resource dev-manabie-online \
 --runtime go116 \
 --build-env-vars-file ./cmd/utils/function/firebase-auth/env_local.yaml \
 --service-account thu-vo@dev-manabie-online.iam.gserviceaccount.com
```

## How to deploy on console

1. Click create new function
2. Select Event Type is `Firebase Auth`
3. Setup environment as env.yaml file in Console https://cloud.google.com/functions/docs/configuring/env-var
4. Include code and deploy
