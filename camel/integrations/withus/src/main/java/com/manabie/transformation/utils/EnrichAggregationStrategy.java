package com.manabie.transformation.utils;

import org.apache.camel.AggregationStrategy;
import org.apache.camel.Exchange;

public class EnrichAggregationStrategy implements AggregationStrategy {

    public Exchange aggregate(Exchange original, Exchange resource) {
        // this is just an example, for real-world use-cases the
        // aggregation strategy would be specific to the use-case

        if (resource == null) {
            return original;
        }
        Object oldBody = original.getIn().getBody(String.class);
        Object newBody = resource.getIn().getBody(String.class);
        original.getIn().setBody(oldBody + ":" + newBody + ":enriched");
        return original;
    }

}