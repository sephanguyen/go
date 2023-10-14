.EXPORT_ALL_VARIABLES:

ENV=local
COMPOSE_PROJECT_NAME=manabie-online
LD_LIBRARY_PATH:=$(LD_LIBRARY_PATH):/usr/local/lib
DOCKER_FILE=./developments/release.Dockerfile
COMPOSE_FILE=./developments/proto.docker-compose.yml
GENPROTO_OUT=/pkg/genproto
TEST_FILE=.
GO_VERSION ?= $(shell cat ./deployments/versions/go)

# Init command, should be run first after cloning the repository to local machines
init:
	git config core.hooksPath .githooks
uninit:
	git config core.hooksPath .git/hooks

# Commands to build server's docker images
build-docker-tools:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/tools:0.0.13 \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg GITHUB_TOKEN=${GITHUB_TOKEN} \
		--file ./developments/tools.Dockerfile .
# Build production server runner image
TAG?=runner
build-docker-runner:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/backend:${TAG} \
		--file ./developments/release.Dockerfile \
		--build-arg GITHUB_TOKEN=${GITHUB_TOKEN} \
		--target runner .
build-docker-dev:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/backend:locally \
		--file ./developments/development.Dockerfile \
		--build-arg GO_VERSION=${GO_VERSION} \
		--target developer .
build-docker-j4:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/backend-j4:${J4_TAG} \
		--file ./developments/development.Dockerfile \
		--build-arg GO_VERSION=${GO_VERSION} \
		--target j4-runner .

build-hasura-migrator:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2-20230411 \
		--file ./developments/release.Dockerfile \
		--target hasura-migrator .

build-hasura-v2-migrator:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v2.8.1.cli-migrations-v3 \
		--file ./developments/release.Dockerfile \
		--target hasura-v2-migrator .

build-docker-imd:
	cd ./developments/import-map-deployer/ && DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/import-map-deployer:0.0.1 \
		-f import-map-deployer.Dockerfile .

build-docker-graphql-mesh:
	cd ./developments/graphql-mesh/ && DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/graphql-mesh:0.0.1 \
		-f graphql-mesh.Dockerfile .

build-docker-kafkatools:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/kafkatools:0.0.2 \
		--file ./developments/kafkatools.Dockerfile \
		--target runner .

build-docker-wait-for:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/wait-for:0.0.2 \
		--file ./developments/waitfor.Dockerfile \
		--target wait-for .

build-docker-mongodb-custom:
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag asia.gcr.io/student-coach-e1e95/mongodb-custom:0.0.1 \
		--file ./developments/mongodb.Dockerfile \
		--target mongodb-custom .

# Command to run unit tests
# When updating this command, remember to update for test-unit-for-coverage as well
test-unit:
	go test -count=3 ./internal/... -cover -covermode=count -coverprofile=cover.out -coverpkg=./internal/...
	go tool cover -func=cover.out -o cover.func
	tail -n 1 cover.func
	go test ./cmd/utils/grafana/...
# This command focuses on getting the coverage, thus tests are run only once.
test-unit-for-coverage:
	go test -count=1 ./internal/... -cover -covermode=count -coverprofile=cover.out.tmp -coverpkg=./internal/...
	cat cover.out.tmp | grep -v "_generated_impl.go" > cover.out && rm cover.out.tmp
	go tool cover -func=cover.out -o cover.func
	tail -n 1 cover.func
test-sqlclosecheck-lint:
	go build -o sqlclosecheck ./cmd/custom_lint/main.go
	-go vet -vettool=./sqlclosecheck ./cmd/custom_lint/testdata 2> sql_close_check_result.txt
	diff -a sql_close_check_result.txt ./cmd/custom_lint/testdata/expected_result.txt
	rm sql_close_check_result.txt
	rm sqlclosecheck
test-helm:
	helm unittest -3 deployments/helm/manabie-all-in-one
	helm unittest -3 deployments/helm/platforms/elastic
	helm unittest -3 deployments/helm/platforms/kafka
	helm unittest -3 deployments/helm/platforms/nats-jetstream
	helm unittest -3 deployments/helm/platforms/unleash
lint:
	golangci-lint run

lint-proto:
	docker compose up check_lint

# Commands to generate various stuffs
# docker compose up syllabus_generate_dart
gen-proto-go:
	docker compose up generate_go
	docker compose up syllabus_generate_go --build
gen-old-proto-dart:
	docker compose up gen_old_dart
gen-proto-dart:
	docker compose up generate_dart
	docker compose up syllabus_generate_dart
gen-proto-ts:
	docker compose up generate_ts
	docker compose up syllabus_generate_ts
gen-proto-ts-v2:
	docker compose up generate_ts_v2
gen-syllabus-proto-ts:
	docker compose up syllabus_generate_ts
gen-syllabus-proto-ts-v2:
	docker compose up syllabus_generate_ts_v2
gen-proto-py:
	docker compose up generate_py
gen-proto:
	docker compose up generate_go
	docker compose up generate_ts
	docker compose up generate_dart

	docker compose up syllabus_generate_go
	docker compose up syllabus_generate_ts
	docker compose up syllabus_generate_dart
syllabus-gen-proto:
	docker compose up syllabus_generate_go --build
	docker compose up syllabus_generate_ts --build
	docker compose up syllabus_generate_dart --build
gen-proto-ts-v3:
	cd developments/bufbuild &&  DOCKER_BUILDKIT=1 docker build --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --progress=plain  --tag bufbuild:0.0.4  -f bufbuild.Dockerfile .
	docker compose up generate_connect_bufbuild
gen-mock-repo:
	go run cmd/utils/main.go mock bob mock/bob
	go run cmd/utils/main.go mock tom mock/tom
	go run cmd/utils/main.go mock eureka mock/eureka
	go run cmd/utils/main.go mock enigma mock/enigma
	go run cmd/utils/main.go mock golibs
	go run cmd/utils/main.go mock draft
	go run cmd/utils/main.go mock zeus
	go run cmd/utils/main.go mock usermgmt mock/usermgmt
	go run cmd/utils/main.go mock payment mock/payment
	go run cmd/utils/main.go mock lessonmgmt mock/lessonmgmt
	go run cmd/utils/main.go mock mastermgmt
	go run cmd/utils/main.go mock entryexitmgmt mock/entryexitmgmt
	go run cmd/utils/main.go mock yasuo mock/yasuo
	go run cmd/utils/main.go mock notification mock/notification
	go run cmd/utils/main.go mock conversationmgmt mock/conversationmgmt
	go run cmd/utils/main.go mock timesheet mock/timesheet
	go run cmd/utils/main.go mock invoicemgmt mock/invoicemgmt
	go run cmd/utils/main.go mock calendar mock/calendar
	go run cmd/utils/main.go mock virtualclassroom mock/virtualclassroom
	go run cmd/utils/main.go mock discount mock/discount
	go run cmd/utils/main.go mock spike mock/spike
	go run cmd/utils/main.go mock auth mock/auth

gen-mock-v2:
	go run cmd/utils/main.go mock eureka_v2

gen-mock-tom-svc:
	mockery  --dir internal/tom/app/core --all --output ./mock/tom/services
	mockery  --dir internal/tom/app/support --all --output ./mock/tom/services
	mockery  --dir internal/tom/app/lesson --all --output ./mock/tom/services

gen-mock-payment-svc:
	mockery  --dir internal/payment/search --case underscore --all --output ./mock/payment/services/search
	mockery  --dir internal/payment/services/domain_service --case underscore --all --output ./mock/payment/services/domain_service
	mockery  --dir internal/payment/services/export_service --case underscore --all --output ./mock/payment/services/export_service
	mockery  --dir internal/payment/services/order_mgmt --case underscore --all --output ./mock/payment/services/order_mgmt
	mockery  --dir internal/payment/services/internal_service --case underscore --all --output ./mock/payment/services/internal_service
	mockery  --dir internal/payment/services/file_service --case underscore --all --output ./mock/payment/services/file_service
	mockery  --dir internal/payment/utils --case underscore --all --output ./mock/payment/utils
	mockery  --dir internal/payment/services/course_mgmt --case underscore --all --output ./mock/payment/services/course_mgmt
	go run cmd/utils/main.go mock payment mock/payment

gen-old-proto:
	docker compose up --exit-code-from gen_old_go

gen-data-pipeline:
	go run cmd/utils/main.go plparser \
	-f ./deployments/helm/platforms/kafka-connect/postgresql2postgresql \
	-t cmd/utils/data_pipeline_parser/pipeline_template.txt \
	-k cmd/utils/data_pipeline_parser/pipeline_source_template.txt \
	-s mock/testing/testdata \
	-o deployments/helm/manabie-all-in-one/charts/hephaestus/generated_connectors/ \
	-o deployments/helm/backend/hephaestus/generated_connectors/

gen-pod-schema:
	./scripts/gen_all_schema.sh
gen-db-schema:
	./scripts/gendbschema.bash

expose-all:
	./scripts/expose_k8s_ports.bash
expose-backend:
	./scripts/expose_k8s_ports.bash backend
expose-emulator:
	./scripts/expose_k8s_ports.bash emulator
expose-appsmith:
	./scripts/expose_k8s_ports.bash appsmith

gen-helm-chart:
	./scripts/template/gen.bash

auto-gen:
	go run developments/generate/main.go
	go fmt ./features/...
	make gen-proto

auto-gen-bdd:
	go run developments/generate/main.go

gen-grafana-dashboard:
	go run cmd/utils/main.go grafana default
	go run cmd/utils/main.go grafana virtualclassroom
	go run cmd/utils/main.go grafana lessonmgmt
	go run cmd/utils/main.go grafana hasura

check-tiertest:
	go run cmd/utils/main.go tiertest --dir ${TEST_DIR} --tier ${TEST_TIER}

gen-rls:
	go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file

gen-rls-stg:
	go run cmd/utils/main.go pg_gen_rls  --rlsType=rls-file --stgHasura=true

samena:
	go install github.com/manabie-com/backend/cmd/samena

.PHONY: update-deps
update-deps:
	./deployments/helm/backend/update_deps.sh
	./deployments/helm/integrations/update_deps.sh
	./camel/integrations/demo/update_deps.sh
	./camel/integrations/quarkus/update_deps.sh
	./deployments/helm/platforms/camel-k/update_deps.sh

sync-chart:
	pip install pyyaml
	python3 scripts/customize_bob_hasura.py
	go run cmd/utils/main.go sync_chart
	make update-deps

sync-chart-e2e:
	go run cmd/utils/main.go sync_chart e2e_local
	./deployments/helm/backend/update_deps.sh

.PHONY: camel
camel:
	kamel run ./camel/integrations/withus/src/main/java/com/manabie/Staging.java \
		-n camel-k --dev
	# kamel run ./camel/integrations/withus/src/main/java/com/manabie/Withus.java \
	# 	--pod-template ./camel/integrations/withus/src/main/resources/deployment.yml \
	# 	-o yaml > ./deployments/helm/integrations/templates/withus.yaml

.PHONY: deploy-camel
deploy-camel:
	# kamel -n camel-k run ./camel/integrations/withus/src/main/java/com/manabie/HelloWorld.java
	kamel -n camel-k run ./camel/integrations/withus/src/main/java/com/manabie/WithusStaging.java
