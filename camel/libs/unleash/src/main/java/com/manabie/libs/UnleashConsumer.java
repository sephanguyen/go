package com.manabie.libs;

import org.apache.camel.Exchange;
import org.apache.camel.Message;
import org.apache.camel.Processor;
import org.apache.camel.support.ScheduledPollConsumer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.concurrent.ExecutorService;

public class UnleashConsumer extends ScheduledPollConsumer {
    private static final Logger LOG = LoggerFactory.getLogger(UnleashConsumer.class);

    private final UnleashEndpoint endpoint;

    private ExecutorService executorService;

    private ManabieUnleash manabieUnleash;

    public UnleashConsumer(UnleashEndpoint endpoint, Processor processor) {
        super(endpoint, processor);
        this.endpoint = endpoint;
        manabieUnleash = new ManabieUnleash(endpoint.getServiceName(), endpoint.getUrl(), endpoint.getApiToken(),
                endpoint.getEnv(), endpoint.getOrg());
    }

    @Override
    protected void doStart() throws Exception {
        super.doStart();
        executorService = endpoint.createExecutor();
        executorService.submit(() -> {
        });
    }

    @Override
    protected void doStop() throws Exception {
        super.doStop();
        getEndpoint().getCamelContext().getExecutorServiceManager().shutdownGraceful(executorService);
    }

    @Override
    protected int poll() throws Exception {
        Exchange exchange = getEndpoint().createExchange();
        Message messageIn = exchange.getIn();
        Boolean isEnabled = manabieUnleash.isEnabled(endpoint.getName());
        LOG.info("unleash producer: " + isEnabled);
        messageIn.setHeader("UNLEASH_ENABLED", isEnabled);

        try {
            getProcessor().process(exchange);
        } catch (Exception e) {
            getExceptionHandler().handleException("Error processing exchange", exchange, e);
        }

        return 1;
    }
}
