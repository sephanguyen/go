// camel-k: trait=pod.enabled=true
// camel-k: dependency=camel:http
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

import javax.sql.DataSource;

import org.apache.camel.BindToRegistry;
import org.apache.camel.PropertyInject;
import org.apache.camel.builder.RouteBuilder;
import org.apache.commons.dbcp.BasicDataSource;
import org.springframework.jdbc.datasource.DataSourceTransactionManager;

public class DBTransactionSQL extends RouteBuilder {

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

        @BindToRegistry("PlatformTransactionManager")
        public DataSourceTransactionManager transactionManager(DataSource dataSource) {
                return new DataSourceTransactionManager(dataSource);
        }

        @Override
        public void configure() throws Exception {
                from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                                .to("direct:projects");

                from("direct:projects")
                                .transacted()
                                .to("direct:insert", "direct:insert2");

                from("direct:insert")
                                .to("sql:insert into t2 values (3, 'test2')")
                                .log("info:${body}");

                from("direct:insert2")
                                .to("sql:insert into t1 values (2, 'test2')")
                                .log("info:${body}");

        }
}