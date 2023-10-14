package com.manabie.libs;

import org.apache.camel.Category;
import org.apache.camel.Consumer;
import org.apache.camel.Processor;
import org.apache.camel.Producer;
import org.apache.camel.spi.Metadata;
import org.apache.camel.spi.UriParam;
import org.apache.camel.spi.UriPath;
import org.apache.camel.support.DefaultEndpoint;
import org.apache.camel.spi.UriEndpoint;

import java.io.IOException;
import java.util.concurrent.ExecutorService;

/**
 * Auth component which does bla bla.
 *
 * TODO: Update one line description above what the component does.
 */
@UriEndpoint(firstVersion = "1.0.0", scheme = "auth", title = "Auth", syntax = "auth:name", category = {
        Category.JAVA})
public class AuthEndpoint extends DefaultEndpoint {

    @UriPath(label = "common", description = "url name")
    @Metadata(required = true)
    private String name;

    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String googleApiKey;

    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String authServiceAddress;
    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String authServicePort;

    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String tenantId;
    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String username;
    @UriParam(label = "common", defaultValue = "", description = "The Object name inside the bucket")
    private String password;

    public AuthEndpoint() {
    }

    public AuthEndpoint(String uri, AuthComponent component, String name) {
        super(uri, component);
        this.setName(name);
    }

    @Override
    public String getEndpointBaseUri() {
        return super.getEndpointBaseUri();
    }

    @Override
    public Producer createProducer() throws Exception {
        return new AuthProducer(this);
    }

    @Override
    public Consumer createConsumer(Processor processor) throws Exception {
        return null;
    }

    public ExecutorService createExecutor() {
        // TODO: Delete me when you implemented your custom component
        return getCamelContext().getExecutorServiceManager().newSingleThreadExecutor(this, "AuthConsumer");
    }

    @Override
    public void close() throws IOException {
        super.close();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getGoogleApiKey() {
        return googleApiKey;
    }

    public void setGoogleApiKey(String googleApiKey) {
        this.googleApiKey = googleApiKey;
    }

    public String getAuthServiceAddress() {
        return authServiceAddress;
    }

    public void setAuthServiceAddress(String authServiceAddress) {
        this.authServiceAddress = authServiceAddress;
    }

    public String getAuthServicePort() {
        return authServicePort;
    }

    public void setAuthServicePort(String authServicePort) {
        this.authServicePort = authServicePort;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }
}
