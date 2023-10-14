.EXPORT_ALL_VARIABLES:

CURRENT_DATE:=$(shell date "+%Y%m%d%H%M%S")
COMMIT_SHA=$(shell git rev-parse --short HEAD)

ENV ?= local
ORG ?= manabie
CAMEL_IMAGE_NAME ?= asia.gcr.io/student-coach-e1e95/camel-demo
CAMEL_IMAGE_TAG ?= local

# Use `make run` to run locally without Docker or K8s
.PHONY: run
run:
	./mvnw package
	./mvnw camel:run

# Use `make build` to build Docker image
.PHONY: build
build: build-cache-if-not-exists
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag $(CAMEL_IMAGE_NAME):$(CAMEL_IMAGE_TAG) \
		--target camel-run \
		--file Dockerfile .

# Use `make deploy` to deploy built Docker image to local K8s.
# Use `make build deploy` to quickly build and deploy in one command.
.PHONY: deploy
deploy:
	ENV=$(ENV) ORG=$(ORG) CAMEL_IMAGE_TAG=$(CAMEL_IMAGE_TAG) \
		skaffoldv2 deploy -f skaffold2.demo.yaml

# Use `make uninstall` to uninstall the demo Helm chart
.PHONY: uninstall
uninstall:
	helm -n $(ENV)-$(ORG)-backend uninstall demo

.PHONY: build-cache
build-cache:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/camel-cache:0.0.1 \
		--target camel-cache \
		--file Dockerfile .

.PHONY: build-cache-if-not-exists
build-cache-if-not-exists:
	if ! docker image inspect asia.gcr.io/student-coach-e1e95/camel-cache:0.0.1 >/dev/null; then \
		make build-cache; \
	fi;
