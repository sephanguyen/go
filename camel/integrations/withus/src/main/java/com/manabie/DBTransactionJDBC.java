// camel-k: trait=pod.enabled=true
// camel-k: dependency=camel:http
// camel-k: dependency=camel:google-storage
// camel-k: dependency=camel:file
// camel-k: dependency=camel:jdbc
// camel-k: dependency=mvn:org.apache.camel.springboot:camel-spring-boot:3.20.1

// camel-k: dependency=mvn:commons-dbcp:commons-dbcp:1.4
// camel-k: dependency=mvn:org.postgresql:postgresql:9.4-1201-jdbc41
// camel-k: dependency=mvn:org.springframework:spring-tx:5.3.10
// camel-k: dependency=mvn:org.springframework:spring-jdbc:5.3.10
// camel-k: service-account=local-camel-k-resource
// camel-k: property=file:../../../resources/application.properties

// Example init db
// create database test1;
// CREATE TABLE t1(a int, b text, PRIMARY KEY(a));
// CREATE TABLE t2(c int, d text, PRIMARY KEY(c));
// insert into t1 values (1, 'test1');
// insert into t2 values (1, 'test1');
// select * from t1;
// select * from t2;

package com.manabie;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.spring.spi.SpringTransactionPolicy;
import org.apache.commons.dbcp.BasicDataSource;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.jdbc.datasource.DataSourceTransactionManager;

public class DBTransactionJDBC extends RouteBuilder {
        @Override
        public void configure() throws Exception {
                BasicDataSource ds = new BasicDataSource();
                ds.setUrl("jdbc:postgresql://postgres-infras.emulator.svc.cluster.local:5432/test1");
                ds.setUsername("postgres");
                ds.setPassword("example");
                ds.setDriverClassName("org.postgresql.Driver");
                from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                                .to("direct:projects");

                PlatformTransactionManager transactionManager = new DataSourceTransactionManager(ds);

                SpringTransactionPolicy propagationRequiredPolicy = new SpringTransactionPolicy();
                propagationRequiredPolicy.setTransactionManager(transactionManager);
                propagationRequiredPolicy.setPropagationBehaviorName("PROPAGATION_REQUIRED");

                // Register transaction policy with Camel context
                getContext().getRegistry().bind("PROPAGATION_REQUIRED", propagationRequiredPolicy);

                getContext().getRegistry().bind("PlatformTransactionManager", transactionManager);
                getContext().getRegistry().bind("myDataSource", ds);

                from("direct:projects")
                                .transacted("PROPAGATION_REQUIRED")
                                .to("direct:insert", "direct:insert2");

                from("direct:insert")
                                .setBody(simple("insert into t1 values (2, 'test1');"))
                                .log("test ${body}")
                                .to("jdbc:myDataSource?useHeadersAsParameters=true")
                                .log("info:${body}");

                from("direct:insert2")
                                .setBody(simple("insert into t2 values (1, 'test1');"))
                                .log("test ${body}")
                                .to("jdbc:myDataSource?useHeadersAsParameters=true")
                                .log("info:${body}");

                // from("direct:projects")
                // .transacted("PROPAGATION_REQUIRED")
                // .to("direct:insert");

                // from("direct:insert")
                // .setBody(simple("insert into t1 values (2, 'test1');insert into t2 values (1,
                // 'test1');"))
                // .log("test ${body}")
                // .to("jdbc:myDataSource?useHeadersAsParameters=true")
                // .log("info:${body}");

        }
}
