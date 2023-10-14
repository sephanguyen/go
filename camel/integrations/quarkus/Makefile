.EXPORT_ALL_VARIABLES:
ENV ?= local
ORG ?= manabie
CAMEL_IMAGE_NAME ?= asia.gcr.io/student-coach-e1e95/camel-quarkus
CAMEL_IMAGE_TAG ?= local

# Use `make run.host` to build and run in host machine.
# Check https://quarkus.io/guides/quarkus-maven-plugin for more information on quarkus maven plugin.
.PHONY: run.host
run.host:
	./mvnw compile quarkus:dev

.PHONY: build
build:
	./mvnw package
	docker build -f src/main/docker/Dockerfile.jvm -t $(CAMEL_IMAGE_NAME):$(CAMEL_IMAGE_TAG) .

# Use `make build.native` to compile and build into a Docker image.
# The resultant image runs in native mode (no JVM).
.PHONY: build.native
build.native:
	./mvnw package -Pnative
	docker build -f src/main/docker/Dockerfile.native-micro -t $(CAMEL_IMAGE_NAME):$(CAMEL_IMAGE_TAG) .

# Use `make deploy` to deploy built Docker image to local K8s.
# Use `make build.native deploy` to quickly build and deploy in one command.
.PHONY: deploy
deploy:
	ENV=$(ENV) ORG=$(ORG) CAMEL_IMAGE_TAG=$(CAMEL_IMAGE_TAG) \
		skaffoldv2 deploy -f skaffold2.quarkus.yaml

# Use `make uninstall` to uninstall the demo Helm chart
.PHONY: uninstall
uninstall:
	helm -n $(ENV)-$(ORG)-backend uninstall quarkus