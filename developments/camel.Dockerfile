# syntax=docker/dockerfile:1.3

FROM eclipse-temurin:17-jdk-jammy

WORKDIR /camel

COPY ./camel/integrations/demo/.mvn camel/integrations/demo/.mvn
COPY ./camel/integrations/demo/mvnw camel/integrations/demo/mvnw
COPY ./camel/integrations/demo/pom.xml camel/integrations/demo/pom.xml

WORKDIR /camel/camel/integrations/demo
RUN ./mvnw dependency:resolve

WORKDIR /camel
COPY ./camel/integrations/demo/src camel/integrations/demo/src

WORKDIR /camel/camel/integrations/demo
RUN ./mvnw package
CMD [ "java", "-jar", "target/demo-1.0-SNAPSHOT-executable-jar.jar" ]
