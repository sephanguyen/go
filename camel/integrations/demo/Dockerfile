# syntax=docker/dockerfile:1.3

FROM eclipse-temurin:20.0.2_9-jdk-jammy AS camel-cache

WORKDIR /camel

RUN curl -LO https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v1.28.0/opentelemetry-javaagent.jar

COPY ./.mvn .mvn
COPY ./mvnw mvnw
COPY ./pom.xml pom.xml
RUN ./mvnw dependency:resolve

COPY ./src src
RUN ./mvnw package

##################
FROM asia.gcr.io/student-coach-e1e95/camel-cache:0.0.1 AS camel-run

WORKDIR /camel

COPY ./.mvn .mvn
COPY ./mvnw mvnw
COPY ./pom.xml pom.xml

COPY ./src src
RUN ./mvnw package

CMD [ "./mvnw", "camel:run", "-Dmaven.test.skip=true" ]
