package com.manabie.libs;

import org.apache.camel.Exchange;
import org.apache.camel.Message;
import org.apache.camel.support.DefaultProducer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class UnleashProducer extends DefaultProducer {
    private static final Logger LOG = LoggerFactory.getLogger(UnleashProducer.class);
    private UnleashEndpoint endpoint;
    private ManabieUnleash manabieUnleash;

    public UnleashProducer(UnleashEndpoint endpoint) {
        super(endpoint);
        this.endpoint = endpoint;
        manabieUnleash = new ManabieUnleash(endpoint.getServiceName(), endpoint.getUrl(), endpoint.getApiToken(),
                endpoint.getEnv(), endpoint.getOrg());
    }

    public void process(Exchange exchange) throws Exception {
        Message messageIn = exchange.getIn();
        Boolean isEnabled = manabieUnleash.isEnabled(endpoint.getName());
        LOG.info("unleash producer: " + isEnabled);
        messageIn.setHeader("UNLEASH_ENABLED", isEnabled);
    }

}
