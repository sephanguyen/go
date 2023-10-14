// camel-k: trait=pod.enabled=true
// camel-k: dependency=camel:saga
// camel-k: dependency=camel:google-storage
// camel-k: dependency=camel:sql

// camel-k: dependency=mvn:commons-dbcp:commons-dbcp:1.4
// camel-k: dependency=mvn:org.postgresql:postgresql:9.4-1201-jdbc41
// camel-k: dependency=mvn:org.springframework:spring-tx:5.3.10
// camel-k: dependency=mvn:org.springframework:spring-jdbc:5.3.10
// camel-k: dependency=mvn:org.springframework.boot:spring-boot-starter-jdbc:3.1.1
// camel-k: dependency=mvn:org.apache.camel.springboot:camel-sql-starter:3.20.1
// camel-k: service-account=local-camel-k-resource
// camel-k: property=file:../../../resources/application.properties
// camel-k: build-property=file:../../../resources/application.properties

// Example init db
// create database test1;
// CREATE TABLE t1(a int, b text, PRIMARY KEY(a));
// CREATE TABLE t2(c int, d text, PRIMARY KEY(c));
// insert into t1 values (1, 'test1');
// insert into t2 values (1, 'test1');
// select * from t1;
// select * from t2;

package com.manabie;

import java.util.UUID;

import javax.sql.DataSource;

import org.apache.camel.BindToRegistry;
import org.apache.camel.Exchange;
import org.apache.camel.PropertyInject;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.model.SagaPropagation;
import org.apache.commons.dbcp.BasicDataSource;

public class DBTransactionSagaMultiComponent extends RouteBuilder {

        @BindToRegistry("dataSource")
        public DataSource dataSource(@PropertyInject("datasource.url") String url,
                        @PropertyInject("datasource.username") String username,
                        @PropertyInject("datasource.password") String password,
                        @PropertyInject("datasource.driverClassName") String driverClassName) {
                BasicDataSource ds = new BasicDataSource();

                ds.setUrl(url);
                ds.setUsername(username);
                ds.setPassword(password);
                ds.setDriverClassName(driverClassName);

                return ds;
        }

        @BindToRegistry("genUUID")
        public UUID genUUID() {
                return UUID.randomUUID();
        }

        @Override
        public void configure() throws Exception {

                getContext().addService(new org.apache.camel.saga.InMemorySagaService());

                from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                                .routeId("timerTriger")
                                .to("direct:projects");

                from("direct:projects")
                                .routeId("call3endpoint")
                                .saga()
                                .propagation(SagaPropagation.REQUIRED)
                                .transform().header(Exchange.SAGA_LONG_RUNNING_ACTION)
                                .to("direct:kafka", "direct:insertt2", "direct:insertt1");

                from("direct:insertt1")
                                // .onException(Exception.class)
                                // .maximumRedeliveries(5).end()
                                .routeId("insertt1")
                                .saga()
                                .propagation(SagaPropagation.MANDATORY)
                                .compensation("direct:rollbackt1")
                                .to("sql:insert into t1 values (1, 'test2')")

                                .to("log:${body}");

                from("direct:insertt2")
                                .routeId("insertt2")
                                .saga()
                                .propagation(SagaPropagation.MANDATORY)
                                .compensation("direct:rollbackt2")
                                .to("sql:insert into t2 values (2, 'test2')")
                                .log("info:${body}");

                from("direct:kafka")
                                .routeId("kafka")
                                .setHeader("kafka.KEY", method(UUID.class, "randomUUID"))
                                .setBody(simple("insert done"))
                                .saga()
                                .propagation(SagaPropagation.MANDATORY)
                                .compensation("direct:kafkarollback")
                                .to("kafka:{{kafka.topic}}?brokers={{kafka.bootstrap-server}}&additional-properties[transactional.id]=#bean:genUUID&additional-properties[enable.idempotence]=true&additional-properties[retries]=5")
                                .log("${body}");

                from("direct:rollbackt1")
                                .routeId("rollbackt1")
                                .transform().header(Exchange.SAGA_LONG_RUNNING_ACTION)
                                .log("running rollbackt1:${body}")
                                .to("sql:delete from t1 where a = 2")
                                .log("running rollbackt1");

                from("direct:rollbackt2")
                                .routeId("rollbackt2")
                                .transform().header(Exchange.SAGA_LONG_RUNNING_ACTION)
                                .log("running rollbackt2:${body}")
                                .to("sql:delete from t2 where c = 2")
                                .log("running rollbackt2");

                from("direct:kafkarollback")
                                .routeId("kafkarollback")
                                .log("info:run kafkarollback saga ${headers} ${body}")
                                .transform().header(Exchange.SAGA_LONG_RUNNING_ACTION)
                                .to("kafka:{{kafka.topic}}?brokers={{kafka.bootstrap-server}}")
                                .onException(Exception.class)
                                .maximumRedeliveries(5)
                                .to("log:kafka saga ${headers} ${body}");

        }
}
