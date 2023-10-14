package com.manabie.transformation;

import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.processor.aggregate.GroupedExchangeAggregationStrategy;

import com.manabie.transformation.utils.Customer;

import org.apache.camel.Exchange;
import org.apache.camel.Processor;

public class TransformByProcessor extends RouteBuilder {

    public class MyProcessor1 implements Processor {
        @Override
        public void process(Exchange exchange) throws Exception {
            int body = exchange.getIn().getBody(Integer.class);
            Customer myClassType = new Customer();
            myClassType.setData(String.valueOf(body));
            myClassType.setId("11");
            exchange.getIn().setBody(myClassType);
        }
    }

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody().constant(1)
                .to("log:info")
                .process(new MyProcessor1())
                .to("log:info")
                .marshal().json()
                .to("log:info")
                .aggregate(new GroupedExchangeAggregationStrategy()).constant(true)
                .completionTimeout(500L)
                .unmarshal().json()
                .to("log:info");
    }
}
