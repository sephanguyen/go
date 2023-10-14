package com.manabie.exception;

import java.io.IOException;

import org.apache.camel.Exchange;
import org.apache.camel.LoggingLevel;
import org.apache.camel.builder.RouteBuilder;

// non handle error
// work as default error handler
// 
public class ErrorHandler extends RouteBuilder {

        @Override
        public void configure() throws Exception {
                // errorHandler(defaultErrorHandler());
                // errorHandler(deadLetterChannel("seda:error-queue"));
                // errorHandler(deadLetterChannel("seda:error-queue").maximumRedeliveries(3));

                // errorHandler(deadLetterChannel("seda:error-queue")
                // .maximumRedeliveries(3)
                // .redeliveryDelay(1000).retryAttemptedLogLevel(LoggingLevel.WARN));

                // onException(IOException.class)
                // .maximumRedeliveries(3)
                // .retryAttemptedLogLevel(LoggingLevel.WARN);

                onException(IOException.class)
                                // .handled(true)
                                .maximumRedeliveries(6)
                                .retryAttemptedLogLevel(LoggingLevel.WARN);

                from("timer://trigger-with-route-exception?fixedRate=true&period=6000")
                                .setBody(simple("timer with id ${uuid}"))
                                .setHeader(Exchange.HTTP_METHOD,
                                                constant(org.apache.camel.component.http.HttpMethods.POST))
                                .log("sending request: ${body}")
                                .to("http://mock-rest-student-import/import-student")
                                .onException(IOException.class).continued(true)
                                // .continued(true)
                                // .retryAttemptedLogLevel(LoggingLevel.WARN)
                                .to("seda:response");

                from("timer://trigger-with-route-no-exception?fixedRate=true&period=6000")
                                .setBody(simple("trigger with id ${uuid}"))
                                .setHeader(Exchange.HTTP_METHOD,
                                                constant(org.apache.camel.component.http.HttpMethods.POST))
                                .log("sending request: ${body}")
                                .to("http://mock-rest-student-import/import-student")
                                .to("seda:response");

                from("timer://trigger-do-try-exception?fixedRate=true&period=6000")
                                .doTry()
                                .setBody(simple("trigger with id ${uuid}"))
                                .setHeader(Exchange.HTTP_METHOD,
                                                constant(org.apache.camel.component.http.HttpMethods.POST))
                                .log("sending request: ${body}")
                                .to("http://mock-rest-student-import/import-student")
                                .to("seda:response")
                                .doCatch(IOException.class)
                                .to("seda:docatch")
                                .doFinally()
                                .to("seda:finall")
                                .end();

                // from("timer://trigger-do-try-exception?fixedRate=true&period=6000")
                // .doTry()
                // .setBody(simple("running with message ${uuid}"))
                // .throwException(new IOException("abcd"))
                // .doCatch(IOException.class).onWhen(exceptionMessage().contains("test"))
                // .to("seda:docatch")
                // .doFinally()
                // .to("seda:finall")
                // .end();

                from("seda:response").log("response: ${body}");
                from("seda:error-queue").log("error: ${body}");
                from("seda:docatch").log("return docatch: ${body}");
                from("seda:finall").log("return finall: ${body}");
        }

}
