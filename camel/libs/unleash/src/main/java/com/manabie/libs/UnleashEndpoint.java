package com.manabie.libs;

import org.apache.camel.Category;
import org.apache.camel.Consumer;
import org.apache.camel.Processor;
import org.apache.camel.Producer;
import org.apache.camel.support.DefaultEndpoint;
import org.apache.camel.spi.Metadata;
import org.apache.camel.spi.UriEndpoint;
import org.apache.camel.spi.UriParam;
import org.apache.camel.spi.UriPath;

import java.util.concurrent.ExecutorService;

/**
 * Unleash component which does bla bla.
 *
 * TODO: Update one line description above what the component does.
 */
@UriEndpoint(firstVersion = "1.0.0", scheme = "unleash", title = "Unleash", syntax = "unleash:name", category = {
        Category.JAVA })
public class UnleashEndpoint extends DefaultEndpoint {
    @UriPath(label = "common", description = "url name")
    @Metadata(required = true)
    private String name;

    @UriParam(label = "common", defaultValue = "manabie", description = "The Object name inside the bucket")
    private String org = "manabie";

    @UriParam(label = "common", defaultValue = "local", description = "The Object name inside the bucket")
    private String env = "local";

    @UriParam(label = "common", description = "The Object name inside the bucket")
    private String serviceName;

    @UriParam(label = "common", description = "The Object name inside the bucket")
    private String url;

    @UriParam(label = "common", description = "The Object name inside the bucket")
    private String apiToken;

    public UnleashEndpoint(String uri, UnleashComponent component, String name) {
        super(uri, component);
        this.setName(name);
    }

    public Producer createProducer() throws Exception {
        return new UnleashProducer(this);
    }

    public Consumer createConsumer(Processor processor) throws Exception {
        Consumer consumer = new UnleashConsumer(this, processor);
        configureConsumer(consumer);
        return consumer;
    }

    /**
     * Some description of this option, and what it does
     */

    public ExecutorService createExecutor() {
        // TODO: Delete me when you implemented your custom component
        return getCamelContext().getExecutorServiceManager().newSingleThreadExecutor(this, "UnleashConsumer");
    }

    public String getUrl() {
        return this.url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public String getOrg() {
        return this.org;
    }

    public void setOrg(String org) {
        this.org = org;
    }

    public String getEnv() {
        return this.env;
    }

    public void setEnv(String env) {
        this.env = env;
    }

    public String getServiceName() {
        return this.serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public String getApiToken() {
        return this.apiToken;
    }

    public void setApiToken(String apiToken) {
        this.apiToken = apiToken;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }

}
