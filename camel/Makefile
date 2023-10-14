integration-test-yaks:
	cd ./integrations/withus && yaks upload bdd/extension/steps -v -n camel-k 
	cd ./integrations/withus && yaks run ./bdd/withus.feature --settings ./yaks-config.yaml -n camel-k --resource ../src/main/java/com/manabie/Withus.java --logger root=DEBUG --property-file application.properties
unit-test:
	cd ./integrations/withus && mvn compile && mvn test
withus:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/Withus.java -t logging.level=INFO -n camel-k
mock-rest-import:
	kamel run --dev ./integrations/withus/src/main/java/com/manabie/MockRestStudentImport.java -n camel-k --dependency file://libs/unleash/target/unleash-1.0.0.jar
consumer-kafka:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/KafkaConsumer.java -t logging.level=INFO -n camel-k
producer-kafka:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/KafkaProducer.java -t logging.level=INFO -n camel-k
unleash-kafka:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/KafkaWithUnleash.java -t logging.level=INFO -n camel-k --dependency file://libs/unleash/target/unleash-1.0.0.jar
transaction:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransaction.java -t logging.level=DEBUG -n camel-k
sql:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransactionSQL.java -t logging.level=DEBUG -n camel-k
sql-kafka:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransactionMultiComponent.java -t logging.level=DEBUG -n camel-k
saga:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransactionSaga.java -t logging.level=DEBUG -n camel-k
saga-multi-compensation:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransactionSagaMultiCompensation.java -t logging.level=DEBUG -n camel-k
saga-multi-conponent:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/DBTransactionSagaMultiComponent.java -t logging.level=INFO -n camel-k
install-testing-tool:
	yaks install -n camel-k --force
integration-test-citrus:
	cd ./integrations/withus && mvn clean verify
integration-test-only:
	cd ./integrations/withus && mvn clean verify -Pintegration-tests
custom-unleash:
	kamel run --dev --pod-template ./integrations/withus/src/main/resources/deployment.yml ./integrations/withus/src/main/java/com/manabie/UnleashExample.java -t logging.level=DEBUG -n camel-k --dependency file://libs/unleash/target/unleash-1.0.0.jar

no-error-example:
	kamel run --dev ./integrations/withus/src/main/java/com/manabie/exception/ErrorHandler.java -t logging.level=INFO -n camel-k
