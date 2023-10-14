## Build a image
Cloud Build (avoid for now, please use local build then push to gcloud registry)
```bash
gcloud builds submit --pack image=gcr.io/student-coach-e1e95/withus-etl
```
## Local Build image in local
Prod and Dev environments sharing the same image, only input arguments are different.
### Withus
```bash
docker build --progress=plain --tag asia.gcr.io/student-coach-e1e95/withus-etl --file ./Dockerfile --target runner .
```
### iTee
```bash
docker build --progress=plain --tag asia.gcr.io/student-coach-e1e95/itee-etl --file ./Dockerfile --target runner .
```

## Push local image to GCloud Registry
### Withus
```bash
docker push asia.gcr.io/student-coach-e1e95/withus-etl
```
### iTee
```bash
docker push asia.gcr.io/student-coach-e1e95/itee-etl
```

## Build a job
```bash
gcloud beta run jobs create withus-etl \
--image asia.gcr.io/student-coach-e1e95/withus-test-job \
--cpu 1 \
--memory 512Mi \
--tasks 1 \
--parallelism 0 \
--max-retries 0 \
--task-timeout 5m \
--region asia-northeast1 \
--vpc-egress all-traffic \
--vpc-connector tokyo \
--set-env-vars SERVER_IP=127.0.0.1,SERVER_PORT=80,USERNAME=test,PASSWORD=test,FILE_PATH=/home/test,FILE_NAME=
```

### Notes:
Job name and region are used to identify a job, **a job name must be unique in a region**.
```bash
gcloud beta run jobs create withus-etl
--vpc-connector tokyo \
```
```bash
gcloud beta run jobs create itee-etl
--vpc-connector tokyo \
```
Always provide minimum resource, to avoid exceeding the free quota limit, also, we don't need too much for current action.
```bash
--cpu 1 \
--memory 512Mi \
--tasks 1 \
```

These arguments are required to use the static ip that whitelisted by Withus.
```bash
--vpc-egress all-traffic \
--vpc-connector tokyo \
```

## Execute a job
### Withus
Dev
```bash
gcloud beta run jobs execute withus-etl-test --region asia-northeast1
```
Prod
```bash
gcloud beta run jobs execute withus-etl --region asia-northeast1
```

### iTee
Dev
```bash
gcloud beta run jobs execute itee-etl-test --region asia-northeast1
```
Prod
```bash
gcloud beta run jobs execute itee-etl --region asia-northeast1
```

## Delete a job
### Withus
Dev
```bash
gcloud beta run jobs delete withus-etl-test --region asia-northeast1
```
Prod
```bash
gcloud beta run jobs delete withus-etl --region asia-northeast1
```

### iTee
Dev
```bash
gcloud beta run jobs delete itee-etl-test --region asia-northeast1
```
Prod
```bash
gcloud beta run jobs delete itee-etl --region asia-northeast1
```
