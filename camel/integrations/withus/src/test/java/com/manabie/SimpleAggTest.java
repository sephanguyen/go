package com.manabie;

import java.util.concurrent.TimeUnit;

import org.apache.camel.RoutesBuilder;
import org.apache.camel.builder.NotifyBuilder;
import org.apache.camel.component.mock.MockEndpoint;
import org.apache.camel.test.junit5.CamelTestSupport;
import org.junit.jupiter.api.Test;

import com.manabie.aggregate.AggCorrelation;
import com.manabie.aggregate.SimpleAgg;

public class SimpleAggTest extends CamelTestSupport {
    @Test
    public void testAggCorrelation() throws Exception {
        NotifyBuilder notification = new NotifyBuilder(context)
                .from("direct:secondRoute").whenDone(1)
                .create();

        notification.matches(2, TimeUnit.SECONDS);

        MockEndpoint.assertIsSatisfied(context);
    }

    @Override
    protected RoutesBuilder createRouteBuilder() {
        return new SimpleAgg();
    }
}