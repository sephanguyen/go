package com.manabie;

import java.util.concurrent.TimeUnit;

import org.apache.camel.RoutesBuilder;
import org.apache.camel.builder.NotifyBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.test.junit5.CamelTestSupport;
import org.junit.jupiter.api.Test;

import com.manabie.aggregate.AggCorrelation;

public class AggCorrelationTest extends CamelTestSupport {
    @Test
    public void testAggCorrelation() throws Exception {
        NotifyBuilder notification = new NotifyBuilder(context)
                .from("direct:mockRoute").whenDone(1)
                .create();

        template.sendBodyAndHeader("direct:start", "A", "correlationID", 1);
        template.sendBodyAndHeader("direct:start", "B", "correlationID", 1);
        template.sendBodyAndHeader("direct:start", "F", "correlationID", 2);
        template.sendBodyAndHeader("direct:start", "C", "correlationID", 1);

        notification.matches(3, TimeUnit.SECONDS);

        MockEndpoint.assertIsSatisfied(context);
    }

    @Override
    protected RoutesBuilder createRouteBuilder() {
        return new AggCorrelation();
    }
}