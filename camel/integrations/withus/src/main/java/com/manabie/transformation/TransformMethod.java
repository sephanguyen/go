package com.manabie.transformation;

import org.apache.camel.Exchange;
import org.apache.camel.Expression;
import org.apache.camel.builder.RouteBuilder;

import com.manabie.transformation.utils.Customer;
import com.manabie.transformation.utils.CustomerParser;

public class TransformMethod extends RouteBuilder {

    @Override
    public void configure() throws Exception {
        from("timer://trigger-get-data-withus?fixedRate=true&period=60000")
                .setBody().constant(1)
                .setHeader("A", constant("B"))
                .to("log:info")
                // .transform().method(CustomerParser.class, "convert")
                // .transform(simple("${body} ${header.A}"))
                .transform(new Expression() {
                    @Override
                    public <T> T evaluate(Exchange exchange, Class<T> type) {
                        Customer myClassType = new Customer();
                        myClassType.setData(String.valueOf(exchange.getIn().getBody()));
                        myClassType.setId("11");
                        return (T) myClassType;
                    }
                })
                .to("log:info")
                .marshal().json()
                .to("log:info")
                .unmarshal().json()
                .to("log:info");
    }
}
